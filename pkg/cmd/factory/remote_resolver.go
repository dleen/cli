package factory

import (
	"errors"
	"sort"

	"github.com/cli/cli/v2/context"
	"github.com/cli/cli/v2/git"
	"github.com/cli/cli/v2/internal/gh"
	"github.com/cli/cli/v2/internal/ghinstance"
	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/pkg/set"
	"github.com/cli/go-gh/v2/pkg/ssh"
)

const (
	GH_HOST = "GH_HOST"
)

// Netflix-specific: map git proxy hostnames to their GitHub Enterprise API host.
// This allows the CLI to work automatically without setting GH_HOST.
var netflixGitProxyMapping = map[string]string{
	"git.netflix.net": "github.netflix.net",
}

type remoteResolver struct {
	readRemotes   func() (git.RemoteSet, error)
	getConfig     func() (gh.Config, error)
	urlTranslator context.Translator
	cachedRemotes context.Remotes
	remotesError  error
}

func (rr *remoteResolver) Resolver() func() (context.Remotes, error) {
	return func() (context.Remotes, error) {
		if rr.cachedRemotes != nil || rr.remotesError != nil {
			return rr.cachedRemotes, rr.remotesError
		}

		gitRemotes, err := rr.readRemotes()
		if err != nil {
			rr.remotesError = err
			return nil, err
		}
		if len(gitRemotes) == 0 {
			rr.remotesError = errors.New("no git remotes found")
			return nil, rr.remotesError
		}

		sshTranslate := rr.urlTranslator
		if sshTranslate == nil {
			sshTranslate = ssh.NewTranslator()
		}
		resolvedRemotes := context.TranslateRemotes(gitRemotes, sshTranslate)

		cfg, err := rr.getConfig()
		if err != nil {
			return nil, err
		}

		authedHosts := cfg.Authentication().Hosts()
		if len(authedHosts) == 0 {
			return nil, errors.New("could not find any host configurations")
		}
		defaultHost, src := cfg.Authentication().DefaultHost()

		// Use set to dedupe list of hosts
		hostsSet := set.NewStringSet()
		hostsSet.AddValues(authedHosts)
		hostsSet.AddValues([]string{defaultHost, ghinstance.Default()})
		hosts := hostsSet.ToSlice()

		// Sort remotes
		sort.Sort(resolvedRemotes)

		rr.cachedRemotes = resolvedRemotes.FilterByHosts(hosts)

		// Filter again by default host if one is set
		// For config file default host fallback to cachedRemotes if none match
		// For environment default host (GH_HOST) do not fallback to cachedRemotes if none match
		if src != "default" {
			filteredRemotes := rr.cachedRemotes.FilterByHosts([]string{defaultHost})
			if isHostEnv(src) || len(filteredRemotes) > 0 {
				rr.cachedRemotes = filteredRemotes
			}
		}

		if len(rr.cachedRemotes) == 0 {
			// Fall back to all remotes for commands that only need owner/repo matching.
			// When GH_HOST is set, override the repo host so API calls go to the correct host.
			// This allows git operations through proxies/bastions where the git remote host
			// differs from the GitHub API host.
			if isHostEnv(src) {
				// Override the repo host with GH_HOST for API operations
				rr.cachedRemotes = overrideRemoteRepoHost(resolvedRemotes, defaultHost)
			} else {
				// Check for Netflix git proxy remotes and auto-map to GitHub Enterprise
				rr.cachedRemotes = applyNetflixProxyMapping(resolvedRemotes)
			}
		}

		return rr.cachedRemotes, nil
	}
}

func isHostEnv(src string) bool {
	return src == GH_HOST
}

// overrideRemoteRepoHost creates new Remote objects with the repo host overridden.
// This is used when GH_HOST is set but no remotes match that host - we keep the
// owner/repo from the git remote but use GH_HOST for API operations.
func overrideRemoteRepoHost(remotes context.Remotes, host string) context.Remotes {
	result := make(context.Remotes, len(remotes))
	for i, r := range remotes {
		result[i] = &context.Remote{
			Remote: r.Remote,
			Repo:   ghrepo.NewWithHost(r.RepoOwner(), r.RepoName(), host),
		}
	}
	return result
}

// applyNetflixProxyMapping checks if any remotes use a known Netflix git proxy
// hostname and maps them to the corresponding GitHub Enterprise API host.
// This allows the CLI to work automatically without setting GH_HOST.
func applyNetflixProxyMapping(remotes context.Remotes) context.Remotes {
	result := make(context.Remotes, len(remotes))
	for i, r := range remotes {
		if apiHost, ok := netflixGitProxyMapping[r.RepoHost()]; ok {
			// Netflix git proxy detected - map to GitHub Enterprise for API calls
			result[i] = &context.Remote{
				Remote: r.Remote,
				Repo:   ghrepo.NewWithHost(r.RepoOwner(), r.RepoName(), apiHost),
			}
		} else {
			// Not a Netflix proxy - keep as-is
			result[i] = r
		}
	}
	return result
}
