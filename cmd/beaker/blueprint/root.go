package blueprint

import (
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/cmd/beaker/image"
	"github.com/allenai/beaker/cmd/beaker/options"
	"github.com/allenai/beaker/config"
)

// NewBlueprintCmd creates the root command for this subpackage.
func NewBlueprintCmd(
	parent *kingpin.Application,
	parentOpts *options.AppOptions,
	config *config.Config,
) {
	o := &image.ImageOptions{AppOptions: parentOpts}
	// TODO message reminding to switch to image commands
	cmd := parent.Command("blueprint", "Manage blueprints")

	cmd.Flag("addr", "Address of the Beaker service.").Default(config.BeakerAddress).StringVar(&o.Addr)

	// Add automatic help generation for the command group.
	var helpSubcommands []string
	cmd.Command("help", "Show help.").Hidden().Default().PreAction(func(c *kingpin.ParseContext) error {
		fullCommand := append([]string{cmd.Model().Name}, helpSubcommands...)
		parent.Usage(fullCommand)
		return nil
	}).Arg("command", "Show help on command.").StringsVar(&helpSubcommands)

	// Attach subcommands.
	newCreateCmd(cmd, o, config)
	newInspectCmd(cmd, o, config)
	newRenameCmd(cmd, o, config)
	newPullCmd(cmd, o, config)
}
