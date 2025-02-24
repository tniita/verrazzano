// Copyright (c) 2022, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package install

import (
	"context"
	"fmt"
	"github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1beta1"
	"github.com/verrazzano/verrazzano/tools/vz/cmd/version"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	cmdhelpers "github.com/verrazzano/verrazzano/tools/vz/cmd/helpers"
	"github.com/verrazzano/verrazzano/tools/vz/pkg/constants"
	"github.com/verrazzano/verrazzano/tools/vz/pkg/helpers"
	"helm.sh/helm/v3/pkg/strvals"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	clipkg "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	CommandName = "install"
	helpShort   = "Install Verrazzano"
	helpLong    = `Install the Verrazzano Platform Operator and install the Verrazzano components specified by the Verrazzano CR provided on the command line`
)

var helpExample = fmt.Sprintf(`
# Install the latest version of Verrazzano using the prod profile. Stream the logs to the console until the install completes.
vz install

# Install version %[1]s using a dev profile, timeout the command after 20 minutes.
vz install --version v%[1]s --set profile=dev --timeout 20m

# Install version %[1]s using a dev profile with kiali disabled and wait for the install to complete.
vz install --version v%[1]s --set profile=dev --set components.kiali.enabled=false

# Install the latest version of Verrazzano using CR overlays and explicit value sets.  Output the logs in json format.
# The overlay files can be a comma-separated list or a series of -f options.  Both formats are shown.
vz install -f base.yaml,custom.yaml --set profile=prod --log-format json
vz install -f base.yaml -f custom.yaml --set profile=prod --log-format json

# Install the latest version of Verrazzano using a Verrazzano CR specified with stdin.
vz install -f - <<EOF
apiVersion: install.verrazzano.io/v1alpha1
kind: Verrazzano
metadata:
  namespace: default
  name: example-verrazzano
EOF`, version.GetCLIVersion())

var logsEnum = cmdhelpers.LogFormatSimple

func NewCmdInstall(vzHelper helpers.VZHelper) *cobra.Command {
	cmd := cmdhelpers.NewCommand(vzHelper, CommandName, helpShort, helpLong)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runCmdInstall(cmd, args, vzHelper)
	}
	cmd.Example = helpExample

	cmd.PersistentFlags().Bool(constants.WaitFlag, constants.WaitFlagDefault, constants.WaitFlagHelp)
	cmd.PersistentFlags().Duration(constants.TimeoutFlag, time.Minute*30, constants.TimeoutFlagHelp)
	cmd.PersistentFlags().String(constants.VersionFlag, constants.VersionFlagDefault, constants.VersionFlagInstallHelp)
	cmd.PersistentFlags().StringSliceP(constants.FilenameFlag, constants.FilenameFlagShorthand, []string{}, constants.FilenameFlagHelp)
	cmd.PersistentFlags().Var(&logsEnum, constants.LogFormatFlag, constants.LogFormatHelp)
	cmd.PersistentFlags().StringArrayP(constants.SetFlag, constants.SetFlagShorthand, []string{}, constants.SetFlagHelp)

	// Initially the operator-file flag may be for internal use, hide from help until
	// a decision is made on supporting this option.
	cmd.PersistentFlags().String(constants.OperatorFileFlag, "", constants.OperatorFileFlagHelp)
	cmd.PersistentFlags().MarkHidden(constants.OperatorFileFlag)

	// Dry run flag is still being discussed - keep hidden for now
	cmd.PersistentFlags().Bool(constants.DryRunFlag, false, "Simulate an install.")
	cmd.PersistentFlags().MarkHidden(constants.DryRunFlag)

	return cmd
}

func runCmdInstall(cmd *cobra.Command, args []string, vzHelper helpers.VZHelper) error {
	// Validate the command options
	err := validateCmd(cmd)
	if err != nil {
		return fmt.Errorf("Command validation failed: %s", err.Error())
	}

	// Get the timeout value for the install command
	timeout, err := cmdhelpers.GetWaitTimeout(cmd)
	if err != nil {
		return err
	}

	// Get the log format value
	logFormat, err := cmdhelpers.GetLogFormat(cmd)
	if err != nil {
		return err
	}

	// Get the kubernetes clientset.  This will validate that the kubeconfig and context are valid.
	kubeClient, err := vzHelper.GetKubeClient(cmd)
	if err != nil {
		return err
	}

	// Get the controller runtime client
	client, err := vzHelper.GetClient(cmd)
	if err != nil {
		return err
	}

	// When --operator-file is not used, get the version from the command line
	var version string
	if !cmd.PersistentFlags().Changed(constants.OperatorFileFlag) {
		version, err = cmdhelpers.GetVersion(cmd, vzHelper)
		if err != nil {
			return err
		}
		fmt.Fprintf(vzHelper.GetOutputStream(), fmt.Sprintf("Installing Verrazzano version %s\n", version))
	}

	// Get the verrazzano install resource to be created
	vz, err := getVerrazzanoYAML(cmd, vzHelper, version)
	if err != nil {
		return err
	}

	// Apply the Verrazzano operator.yaml.
	lastTransitionTime := metav1.Now()
	err = cmdhelpers.ApplyPlatformOperatorYaml(cmd, client, vzHelper, version)
	if err != nil {
		return err
	}

	// Wait for the platform operator to be ready before we create the Verrazzano resource.
	vpoPodName, err := cmdhelpers.WaitForPlatformOperator(client, vzHelper, v1beta1.CondInstallComplete, lastTransitionTime)
	if err != nil {
		return err
	}

	// Create the Verrazzano install resource.
	// We will retry up to 5 times if there is an error.
	// Sometimes we see intermittent webhook errors due to timeouts.
	retry := 0
	for {
		err = client.Create(context.TODO(), vz)
		if err != nil {
			if retry == 5 {
				return fmt.Errorf("Failed to create the verrazzano install resource: %s", err.Error())
			}
			time.Sleep(time.Second)
			retry++
			fmt.Fprintf(vzHelper.GetOutputStream(), fmt.Sprintf("Retrying after failing to create the verrazzano install resource: %s\n", err.Error()))
			continue
		}
		break
	}

	// Wait for the Verrazzano install to complete
	return waitForInstallToComplete(client, kubeClient, vzHelper, vpoPodName, types.NamespacedName{Namespace: vz.GetNamespace(), Name: vz.GetName()}, timeout, logFormat)
}

// getVerrazzanoYAML returns the verrazzano install resource to be created
func getVerrazzanoYAML(cmd *cobra.Command, vzHelper helpers.VZHelper, version string) (vz clipkg.Object, err error) {
	// Get the list yaml filenames specified
	filenames, err := cmd.PersistentFlags().GetStringSlice(constants.FilenameFlag)
	if err != nil {
		return nil, err
	}

	// Get the set arguments - returning a list of properties and value
	pvs, err := getSetArguments(cmd, vzHelper)
	if err != nil {
		return nil, err
	}

	// If no yamls files were passed on the command line then create a minimal verrazzano
	// resource.  The minimal resource is used to create a resource called verrazzano
	// in the default namespace using the prod profile.
	var gv schema.GroupVersion
	if len(filenames) == 0 {
		gv, vz, err = helpers.NewVerrazzanoForVZVersion(version)
		if err != nil {
			return nil, err
		}
	} else {
		// Merge the yaml files passed on the command line
		obj, err := cmdhelpers.MergeYAMLFiles(filenames, os.Stdin)
		if err != nil {
			return nil, err
		}
		gv = obj.GroupVersionKind().GroupVersion()
		vz = obj
	}

	// Generate yaml for the set flags passed on the command line
	outYAML, err := generateYAMLForSetFlags(pvs)
	if err != nil {
		return nil, err
	}

	// Merge the set flags passed on the command line. The set flags take precedence over
	// the yaml files passed on the command line.
	vz, err = cmdhelpers.MergeSetFlags(gv, vz, outYAML)
	if err != nil {
		return nil, err
	}

	// Return the merged verrazzano install resource to be created
	return vz, nil
}

// generateYAMLForSetFlags creates a YAML string from a map of property value pairs representing --set flags
// specified on the install command
func generateYAMLForSetFlags(pvs map[string]string) (string, error) {
	yamlObject := map[string]interface{}{}
	for path, value := range pvs {
		// replace unwanted characters in the value to avoid splitting
		ignoreChars := ",[.{}"
		for _, char := range ignoreChars {
			value = strings.Replace(value, string(char), "\\"+string(char), -1)
		}

		composedStr := fmt.Sprintf("%s=%s", path, value)
		err := strvals.ParseInto(composedStr, yamlObject)
		if err != nil {
			return "", err
		}
	}

	yamlFile, err := yaml.Marshal(yamlObject)
	if err != nil {
		return "", err
	}

	yamlString := string(yamlFile)

	// Replace any double-quoted strings that are surrounded by single quotes.
	// These type of strings are problematic for helm.
	yamlString = strings.ReplaceAll(yamlString, "'\"", "\"")
	yamlString = strings.ReplaceAll(yamlString, "\"'", "\"")

	return yamlString, nil
}

// getSetArguments gets all the set arguments and returns a map of property/value
func getSetArguments(cmd *cobra.Command, vzHelper helpers.VZHelper) (map[string]string, error) {
	setMap := make(map[string]string)
	setFlags, err := cmd.PersistentFlags().GetStringArray(constants.SetFlag)
	if err != nil {
		return nil, err
	}

	invalidFlag := false
	for _, setFlag := range setFlags {
		pv := strings.Split(setFlag, "=")
		if len(pv) != 2 {
			fmt.Fprintf(vzHelper.GetErrorStream(), fmt.Sprintf("Invalid set flag \"%s\" specified. Flag must be specified in the format path=value\n", setFlag))
			invalidFlag = true
			continue
		}
		if !invalidFlag {
			path, value := strings.TrimSpace(pv[0]), strings.TrimSpace(pv[1])
			if !strings.HasPrefix(path, "spec.") {
				path = "spec." + path
			}
			setMap[path] = value
		}
	}

	if invalidFlag {
		return nil, fmt.Errorf("Invalid set flag(s) specified")
	}

	return setMap, nil
}

// waitForInstallToComplete waits for the Verrazzano install to complete and shows the logs of
// the ongoing Verrazzano install.
func waitForInstallToComplete(client clipkg.Client, kubeClient kubernetes.Interface, vzHelper helpers.VZHelper, vpoPodName string, namespacedName types.NamespacedName, timeout time.Duration, logFormat cmdhelpers.LogFormat) error {
	return cmdhelpers.WaitForOperationToComplete(client, kubeClient, vzHelper, vpoPodName, namespacedName, timeout, logFormat, v1beta1.CondInstallComplete)
}

// validateCmd - validate the command line options
func validateCmd(cmd *cobra.Command) error {
	if cmd.PersistentFlags().Changed(constants.VersionFlag) && cmd.PersistentFlags().Changed(constants.OperatorFileFlag) {
		return fmt.Errorf("--%s and --%s cannot both be specified", constants.VersionFlag, constants.OperatorFileFlag)
	}
	return nil
}
