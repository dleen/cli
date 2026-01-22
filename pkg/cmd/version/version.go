package version

import (
	"fmt"
	"strings"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewCmdVersion(f *cmdutil.Factory, version, buildDate string) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "version",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprint(f.IOStreams.Out, cmd.Root().Annotations["versionInfo"])
		},
	}

	cmdutil.DisableAuthCheck(cmd)

	return cmd
}

func Format(version, buildDate string) string {
	version = strings.TrimPrefix(version, "v")

	var dateStr string
	if buildDate != "" {
		dateStr = fmt.Sprintf(" (%s)", buildDate)
	}

	return fmt.Sprintf("gh version %s%s [Netflix]\n%s\n", version, dateStr, changelogURL(version))
}

func changelogURL(version string) string {
	// Netflix fork releases
	return "https://github.netflix.net/dleen/cli/releases"
}
