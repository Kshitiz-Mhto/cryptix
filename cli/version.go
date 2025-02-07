package cli

import (
	"fmt"
	"runtime"

	"github.com/Kshitiz-Mhto/cryptix/pkg/env"
	"github.com/Kshitiz-Mhto/cryptix/utility"
	"github.com/spf13/cobra"
)

const logo = `
 ██████╗██████╗ ██╗   ██╗██████╗ ████████╗██╗██╗  ██╗
██╔════╝██╔══██╗╚██╗ ██╔╝██╔══██╗╚══██╔══╝██║╚██╗██╔╝
██║     ██████╔╝ ╚████╔╝ ██████╔╝   ██║   ██║ ╚███╔╝ 
██║     ██╔══██╗  ╚██╔╝  ██╔═══╝    ██║   ██║ ██╔██╗ 
╚██████╗██║  ██║   ██║   ██║        ██║   ██║██╔╝ ██╗
 ╚═════╝╚═╝  ╚═╝   ╚═╝   ╚═╝        ╚═╝   ╚═╝╚═╝  ╚═╝
                                                     
`

var (
	quiet      bool
	verbose    bool
	versionCMD = &cobra.Command{
		Use:   "version",
		Short: "Version will output the current build information",
		Run: func(cmd *cobra.Command, args []string) {
			buildDate := utility.GetBuildDate()
			switch {
			case verbose:
				fmt.Print(logo)
				fmt.Printf("CLI version: v%s\n", env.Vars.CLI_VERSION)
				fmt.Printf("Go version (client): %s\n", runtime.Version())
				if buildDate != "" {
					fmt.Printf("Build date (client): %s\n", buildDate)
				}
				fmt.Printf("OS/Arch (client): %s/%s\n", runtime.GOOS, runtime.GOARCH)
			case quiet:
				fmt.Printf("v%s\n", env.Vars.CLI_VERSION)
			default:
				fmt.Printf("dSync CLI v%s\n", env.Vars.CLI_VERSION)
			}
		},
	}
)

func init() {
	versionCMD.Flags().BoolVarP(&quiet, "quiet", "q", false, "Use quiet output for simple output")
	versionCMD.Flags().BoolVarP(&verbose, "verbose", "v", false, "Use verbose output to see full information")
}
