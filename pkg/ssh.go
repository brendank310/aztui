package ssh

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"github.com/google/uuid"

	armruntime "github.com/Azure/azure-sdk-for-go/sdk/azcore/arm/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/serialconsole/armserialconsole"
)

// Generated from example definition: https://github.com/Azure/azure-rest-api-specs/tree/main/specification/compute/resource-manager/Microsoft.Compute/ComputeRP/stable/2022-08-01/examples/sshPublicKeyExamples/SshPublicKey_Get.json
func AzurePublicKeyGet(subscriptionID string, resourceGroup string, vmName string) ([]string, error) {
	pubKeys := []string{}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return pubKeys, err
	}
	ctx := context.Background()
	client, err := armcompute.NewVirtualMachinesClient(subscriptionID, cred, nil)
	if err != nil {
		return pubKeys, err
	}

	res, err := client.Get(ctx, resourceGroup, vmName, nil)
	if err != nil {
		return pubKeys, err
	}

	for _, pubkey := range res.Properties.OSProfile.LinuxConfiguration.SSH.PublicKeys {
		pubKeys = append(pubKeys, *pubkey.KeyData)
	}

	return pubKeys, nil
}

func GenerateRandomString(size uint) (string, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return strings.Trim(base32.StdEncoding.EncodeToString(b), "="), nil
}

func PublishUserCredentials(subscriptionID string,
	resourceGroupName string,
	tenantID string,
	objectID string,
	location string,
	vmName string,
	user string,
	password string) (*url.URL, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	// Create resource group
	_, err = resourceGroupClient.CreateOrUpdate(context.TODO(),
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.Ptr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}

	// Create keyvault to store user credentials
	kvClient, err := armkeyvault.NewVaultsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := kvClient.BeginCreateOrUpdate(
		context.TODO(),
		resourceGroupName,
		vmName+"-creds",
		armkeyvault.VaultCreateOrUpdateParameters{
			Location: to.Ptr(location),
			Properties: &armkeyvault.VaultProperties{
				AccessPolicies: []*armkeyvault.AccessPolicyEntry{
					{
						ObjectID: to.Ptr(objectID),
						Permissions: &armkeyvault.Permissions{
							Secrets: []*armkeyvault.SecretPermissions{
								to.Ptr(armkeyvault.SecretPermissionsSet),
								to.Ptr(armkeyvault.SecretPermissionsGet),
								to.Ptr(armkeyvault.SecretPermissionsList),
								to.Ptr(armkeyvault.SecretPermissionsDelete),
								to.Ptr(armkeyvault.SecretPermissionsPurge),
							},
						},
						TenantID: to.Ptr(tenantID),
					},
				},
				SKU: &armkeyvault.SKU{
					Name:   to.Ptr(armkeyvault.SKUNameStandard),
					Family: to.Ptr(armkeyvault.SKUFamilyA),
				},
				TenantID: to.Ptr(tenantID),
			},
		},
		nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	// Create secret in keyvault to store the user credentials
	client, err := azsecrets.NewClient(*resp.Vault.Properties.VaultURI, cred, nil)
	if err != nil {
		return nil, err
	}

	secret := fmt.Sprintf("username: %v\npassword: %v\n", user, password)
	log.Printf("Vault name: %v Vault URI: %v", *resp.Vault.Name, *resp.Vault.Properties.VaultURI)
	_, err = client.SetSecret(context.TODO(), *resp.Vault.Name, azsecrets.SetSecretParameters{Value: &secret}, nil)
	if err != nil {
		return nil, err
	}

	uri, err := url.Parse(*resp.Vault.Properties.VaultURI)
	if err != nil {
		return nil, err
	}

	return uri, nil
}

func updateUserCredentials(subscriptionID string, resourceGroupName string, virtualMachineName string, updateProtectedSettings map[string]string) error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	extensionClient, err := armcompute.NewVirtualMachineExtensionsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	// Get the current enablevmaccess properties
	extension, err := extensionClient.Get(context.TODO(), resourceGroupName, virtualMachineName, "enablevmaccess", nil)
	if err != nil {
		return err
	}

	_, err = extensionClient.BeginUpdate(context.TODO(), resourceGroupName, virtualMachineName, "enablevmaccess", armcompute.VirtualMachineExtensionUpdate{
		Properties: &armcompute.VirtualMachineExtensionUpdateProperties{
			Type:                    extension.Properties.Type,
			AutoUpgradeMinorVersion: extension.Properties.AutoUpgradeMinorVersion,
			Publisher:               extension.Properties.Publisher,
			ProtectedSettings:       updateProtectedSettings,
		}}, nil)
	if err != nil {
		return err
	}

	return nil
}

func SetTemporaryUserCredentials(subscriptionID string, resourceGroupName string, virtualMachineName string, username string, password string) error {
	return updateUserCredentials(subscriptionID, resourceGroupName, virtualMachineName, map[string]string{
		"username": username,
		"password": password,
	})
}

func RemoveTemporaryUser(subscriptionID string, resourceGroupName string, virtualMachineName string, username string) error {
	return updateUserCredentials(subscriptionID, resourceGroupName, virtualMachineName, map[string]string{
		"remove_user": username,
	})
}

func SendReset(token string, connURL string) {
	adminURL := strings.Replace(connURL, "/client", "/adminCommand/reset", 1)
	adminURL = strings.Replace(adminURL, "wss://", "https://", 1)
	uuid := uuid.New()
	sysRq := fmt.Sprintf(`{"command":"reset", "requestId": "%v", "commandParameters": {}}`, uuid.String())

	req, err := http.NewRequest("POST", adminURL, bytes.NewBuffer([]byte(sysRq)))
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
	req.Close = true

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
	} else {
		defer resp.Body.Close()
		_, _ = ioutil.ReadAll(resp.Body)
	}
}

func StartSerialConsole(subscriptionID string, resourceGroupName string, virtualMachineName string) (string, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return "", err
	}

	// It looks like this SDK is broken because it is generated based on our
	// busted swagger specs which are incorrect
	// ---
	// So here's the nasty ugly way of doing it, rather than a nice
	// armserialclient.NewClient(...).Create(...)
	vmResourceID := fmt.Sprintf(
		"/subscriptions/%v/resourcegroups/%v/providers/Microsoft.Compute/%v/%v",
		subscriptionID,
		resourceGroupName,
		"virtualMachines",
		virtualMachineName)

	serialPortResource := fmt.Sprintf(
		"/providers/%v/serialPorts/%v/connect",
		"Microsoft.SerialConsole",
		"0")

	armURL := "https://management.azure.com"

	pipeline, err := armruntime.NewPipeline("module", "version", cred,
		runtime.PipelineOptions{}, nil)
	if err != nil {
		return "", err
	}

	endpoint := armURL + vmResourceID + serialPortResource
	req, err := runtime.NewRequest(context.Background(), http.MethodPost,
		endpoint)
	if err != nil {
		return "", err
	}

	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2018-05-01")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header["Accept"] = []string{"application/json"}
	state := armserialconsole.SerialPortStateEnabled
	err = runtime.MarshalAsJSON(req, armserialconsole.SerialPort{
		Properties: &armserialconsole.SerialPortProperties{
			State: &state,
		},
	})

	if err != nil {
		return "", err
	}

	res, err := pipeline.Do(req)
	if err != nil {
		return "", err
	}

	buf := make([]byte, 4096)
	n, err := res.Body.Read(buf)
	if err != nil {
		return "", err
	}

	type ConnectResponse struct {
		ConnectionString string `json:"connectionString"`
	}

	connRes := ConnectResponse{}
	err = json.Unmarshal(buf[:n], &connRes)
	if err != nil {
		return "", err
	}

	if connRes.ConnectionString == "" {
		return "", fmt.Errorf("empty connection string")
	}

	return connRes.ConnectionString + "?authorization=" +
		strings.TrimPrefix(res.Request.Header.Get("Authorization"), "Bearer "), err
}
