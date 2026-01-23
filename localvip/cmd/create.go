package cmd

import (
	"log/slog"
	"os"
	"slices"
	"strings"

	iptables "github.com/VITObelgium/fakes3pp/localvip/iptables"
	nftables "github.com/VITObelgium/fakes3pp/localvip/nftables"
	"github.com/spf13/cobra"
)

const argVipAddress string = "vip-address"
const argTargetAddress string = "target-address"
const argPort string = "port"
const argBackend string = "backend"

const backendIptables string = "iptables"
const backendNftables string = "nftables"

var possibleBackends []string = []string{backendIptables, backendNftables}

const envTargetAddress string = "LOCALVIP_TARGET_ADDRESS"
const envVipAddress string = "LOCALVIP_VIP_ADDRESS"
const envLocalVipBackend string = "LOCALVIP_BACKEND"

const exitCodeGenericFailure = 1
const exitCodeMissingRequiredArgument = 2
const exitCodeInvalidArgument = 3

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Allow creation of a virtual IP address",
	Long: `Allow creation of a virtual IP address that will NAT traffic to another target address.

You can specify multiple port pairs.`,
	PreRun: setupLogging,
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("create called")
		getStringArg := func(name, env_var_fallback string) string {
			return getStringArg(cmd, name, env_var_fallback)
		}
		getStringSlicesArg := func(name string) []string {
			return getStringSlicesArg(cmd, name)
		}
		vipAddress := getStringArg(argVipAddress, envVipAddress)
		targetAddress := getStringArg(argTargetAddress, envTargetAddress)
		ports := getStringSlicesArg(argPort)
		backend := getStringArg(argBackend, envLocalVipBackend)
		if !slices.Contains(possibleBackends, backend) {
			slog.Error("Invalid backend provided", "provided", backend, "allowed", possibleBackends)
			os.Exit(exitCodeInvalidArgument)
		}

		for _, port := range ports {
			slog.Debug("Processing port config", "port", port)
			var vipPort string
			var targetPort string
			portParts := strings.Split(port, ":")
			if len(portParts) == 2 {
				slog.Debug("Portmapping provided", "vip-port", portParts[0], "target-port", portParts[1])
				vipPort = portParts[0]
				targetPort = portParts[1]
			} else if len(portParts) == 1 {
				slog.Debug("Single port provided", "vip-port", portParts[0], "target-port", portParts[1])
				vipPort = portParts[0]
				targetPort = portParts[0]
			} else {
				slog.Error("Invalid value for ports; must be 1 or 2 colon-separated port numbers", "given-value", "port")
				os.Exit(1)
			}
			slog.Debug("Trying to setup NAT", "vipAddress", vipAddress, "vipPort", vipPort, "target-address", targetAddress, "target-port", targetPort)
			var err error
			switch backend {
			case "iptables":
				err = iptables.CreateVip(vipAddress, vipPort, targetAddress, targetPort)
			case backendNftables:
				err = nftables.CreateVip(vipAddress, vipPort, targetAddress, targetPort)

			}

			if err != nil {
				slog.Error("Error encountered when trying to configure iptables", "error", err)
				os.Exit(exitCodeGenericFailure)
			}
			slog.Info("Finished setup NAT", "vipAddress", vipAddress, "vipPort", vipPort, "target-address", targetAddress, "target-port", targetPort)
		}
	},
}

// get a string Argument by name
// If it is not present fallback to the corresponding env_far_fallback
// or if that is not available then error out
func getStringArg(cmd *cobra.Command, name, env_var_fallback string) string {
	v, err := cmd.Flags().GetString(name)
	if err != nil {
		slog.Error("Issue getting argument", "argument-name", name, "error", err)
		os.Exit(exitCodeInvalidArgument)
	}
	if v != "" {
		slog.Debug("Retrieved string argument", "name", name, "value", v, "source", "cli-argument")
		return v
	}
	if env_var_fallback == "" {
		slog.Error("Missing argument value", "argument-name", name)
		os.Exit(exitCodeMissingRequiredArgument)
	}

	v = os.Getenv(env_var_fallback)
	if v == "" {
		slog.Error("Missing argument", "argument-name", name, "env-var-name", env_var_fallback)
		os.Exit(exitCodeMissingRequiredArgument)
	}
	slog.Debug("Retrieved string argument", "name", name, "value", v, "source", "env", "env-var-name", env_var_fallback)
	return v
}

func getStringSlicesArg(cmd *cobra.Command, name string) []string {
	v, err := cmd.Flags().GetStringSlice(name)
	if err != nil {
		slog.Error("Issue getting argument", "argument-name", name, "error", err)
		os.Exit(exitCodeInvalidArgument)
	}
	if len(v) == 0 {
		slog.Error("Missing argument value", "argument-name", name)
		os.Exit(exitCodeMissingRequiredArgument)
	}
	return v
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	createCmd.Flags().String(argVipAddress, "169.254.83.51", "The vitual address that must be mapped to another address")
	createCmd.Flags().String(argTargetAddress, "", "The address to be reached")
	createCmd.Flags().String(argBackend, "iptables", "The backend used for implementing NAT rules {iptables|nftables}")
	createCmd.Flags().StringSlice(argPort, []string{}, `The port to be reached.
	This flag can be used multiple times if multiple ports must be reachable.
	It is also allowed to specify a port pair separated by a colon.
	In that case the first port is for the virtual address and it goes to the 2nd port number on the target address`)

}
