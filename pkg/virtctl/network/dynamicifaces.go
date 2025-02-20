/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2023 Red Hat, Inc.
 *
 */

package network

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"k8s.io/client-go/tools/clientcmd"

	v1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/kubecli"

	"kubevirt.io/kubevirt/pkg/virtctl/templates"
)

const (
	HotplugCmdName = "addinterface"

	ifaceNameArg   = "iface-name"
	networkNameArg = "network-name"
)

var (
	ifaceName                       string
	networkAttachmentDefinitionName string
	persist                         bool
)

type dynamicIfacesCmd struct {
	kvClient     kubecli.KubevirtClient
	isPersistent bool
	namespace    string
}

func NewAddInterfaceCommand(clientConfig clientcmd.ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "addinterface VM",
		Short:   "add a network interface to a running VM",
		Example: usageAddInterface(),
		Args:    templates.ExactArgs(HotplugCmdName, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newDynamicIfaceCmd(clientConfig, persist)
			if err != nil {
				return fmt.Errorf("error creating the `AddInterface` command: %w", err)
			}
			return c.addInterface(args[0], networkAttachmentDefinitionName, ifaceName)
		},
	}
	cmd.SetUsageTemplate(templates.UsageTemplate())
	cmd.Flags().StringVar(&networkAttachmentDefinitionName, networkNameArg, "", "The referenced network-attachment-definition name. Format:\n<netName>, <ns>/<netName>")
	_ = cmd.MarkFlagRequired(networkNameArg)
	cmd.Flags().StringVar(&ifaceName, ifaceNameArg, "", "Logical name of the interface to be plugged")
	_ = cmd.MarkFlagRequired(ifaceNameArg)
	cmd.Flags().BoolVar(&persist, "persist", false, "When set, the added interface will be persisted in the VM spec (if it exists)")

	return cmd
}

func usageAddInterface() string {
	usage := `  #Dynamically attach a network interface to a running VM.
  {{ProgramName}} addinterface <vmi-name> --network-name <net name> --iface-name <iface name>

  #Dynamically attach a network interface to a running VM and persisting it in the VM spec. At next VM restart the network interface will be attached like any other network interface.
  {{ProgramName}} addinterface <vm-name> --network-name <net name> --iface-name <iface name> --persist
  `
	return usage
}

func newDynamicIfaceCmd(clientCfg clientcmd.ClientConfig, persistState bool) (*dynamicIfacesCmd, error) {
	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("cannot obtain KubeVirt client: %v", err)
	}
	namespace, _, err := clientCfg.Namespace()
	if err != nil {
		return nil, err
	}
	return &dynamicIfacesCmd{kvClient: virtClient, isPersistent: persistState, namespace: namespace}, nil
}

func (dic *dynamicIfacesCmd) addInterface(vmName string, networkName string, ifaceName string) error {
	if dic.isPersistent {
		return dic.kvClient.VirtualMachine(dic.namespace).AddInterface(
			context.Background(),
			vmName,
			&v1.AddInterfaceOptions{
				NetworkName:   networkName,
				InterfaceName: ifaceName,
			},
		)
	}
	return dic.kvClient.VirtualMachineInstance(dic.namespace).AddInterface(
		context.Background(),
		vmName,
		&v1.AddInterfaceOptions{
			NetworkName:   networkName,
			InterfaceName: ifaceName,
		},
	)
}
