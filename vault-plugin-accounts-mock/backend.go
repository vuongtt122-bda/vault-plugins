package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/jsonutil"
	"github.com/hashicorp/vault/sdk/logical"
)

// backend wraps the backend framework and adds a map for storing key value pairs
type backend struct {
	*framework.Backend
}

type Account struct {
	AccountId string
	PublicKey string
	PrivateKey string
}

var _ logical.Factory = Factory

// Factory configures and returns Mock backends
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b, err := newBackend()
	if err != nil {
		return nil, err
	}

	if conf == nil {
		return nil, fmt.Errorf("configuration passed into backend is nil")
	}

	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}

	return b, nil
}

func newBackend() (*backend, error) {
	b := &backend{}

	b.Backend = &framework.Backend{
		Help:        strings.TrimSpace(mockHelp),
		BackendType: logical.TypeLogical,
		Paths: framework.PathAppend(
			b.accountPaths(),
			b.signPaths(),
		),
	}

	return b, nil
}

func (b *backend) signPaths() []*framework.Path {
	return []*framework.Path{
		{
			Pattern: "sign",

			Fields: map[string]*framework.FieldSchema{
				"message": {
					Type:        framework.TypeString,
					Description: "Specifies the path of the secret.",
				},
				"accountId": {
					Type:        framework.TypeString,
					Description: "Specifies the accountId of the secret.",
				},
			},

			Operations: map[logical.Operation]framework.OperationHandler{
				logical.ReadOperation: &framework.PathOperation{
					Callback: b.handleReadSign,
					Summary:  "Retrieve the secret from the map.",
				},
			},

			ExistenceCheck: b.handleExistenceCheckSign,
		},
	}
}

func (b *backend) handleExistenceCheckSign(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	out, err := req.Storage.Get(ctx, req.Path)
	if err != nil {
		return false, errwrap.Wrapf("existence check failed: {{err}}", err)
	}

	return out != nil, nil
}

func (b *backend) handleReadSign(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	if req.ClientToken == "" {
		return nil, fmt.Errorf("client token empty")
	}

	accountId := data.Get("accountId").(string)
	message := data.Get("message").(string)

	// Decode the data
	var rawData Account
	entry, err := req.Storage.Get(ctx, req.ClientToken + "/" + accountId) // ignore: req.ClientToken + "/" + accountId
	if err != nil {
		return nil, err
	}
	fetchedData := entry.Value
	if fetchedData == nil {
		resp := logical.ErrorResponse("No value at %v%v", req.MountPoint, accountId)
		return resp, nil
	}

	if err := jsonutil.DecodeJSON(fetchedData, &rawData); err != nil {
		return nil, errwrap.Wrapf("json decoding failed: {{err}}", err)
	}


	b.Logger().Info(fmt.Sprintf("\n\n %v \n\n", rawData.PrivateKey))

	signatureEncode64 := message + "-" + rawData.PrivateKey + "-" + rawData.PublicKey
	resData := map[string]interface{}{
		"signatureEncode64": signatureEncode64,
	}
	// Generate the response
	resp := &logical.Response{
		Data: resData,
	}

	return resp, nil
}

func (b *backend) accountPaths() []*framework.Path {
	return []*framework.Path{
		{
			Pattern: "account",

			Fields: map[string]*framework.FieldSchema{
				// "path": {
				// 	Type:        framework.TypeString,
				// 	Description: "Specifies the path of the secret.",
				// },
				"accountId": {
					Type:        framework.TypeString,
					Description: "Specifies the accountId of the secret.",
				},
			},

			Operations: map[logical.Operation]framework.OperationHandler{
				logical.ReadOperation: &framework.PathOperation{
					Callback: b.handleReadAccount,
					Summary:  "Retrieve the secret from the map.",
				},
				logical.CreateOperation: &framework.PathOperation{
					Callback: b.handleWriteAccount,
				},
			},

			ExistenceCheck: b.handleExistenceCheckAccount,
		},
	}
}

func (b *backend) handleExistenceCheckAccount(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	out, err := req.Storage.Get(ctx, req.Path)
	if err != nil {
		return false, errwrap.Wrapf("existence check failed: {{err}}", err)
	}

	return out != nil, nil
}

func (b *backend) handleReadAccount(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	if req.ClientToken == "" {
		return nil, fmt.Errorf("client token empty")
	}

	accountId := data.Get("accountId").(string)

	// Decode the data
	var rawData Account
	entry, err := req.Storage.Get(ctx, req.ClientToken + "/" + accountId) // ignore: req.ClientToken + "/" + accountId
	if err != nil {
		return nil, err
	}
	fetchedData := entry.Value
	if fetchedData == nil {
		resp := logical.ErrorResponse("No value at %v%v", req.MountPoint, accountId)
		return resp, nil
	}

	if err := jsonutil.DecodeJSON(fetchedData, &rawData); err != nil {
		return nil, errwrap.Wrapf("json decoding failed: {{err}}", err)
	}


	b.Logger().Info(fmt.Sprintf("\n\n %v \n\n", rawData.PrivateKey))

	signatureEncode64 := rawData.PrivateKey + "-" + rawData.PublicKey
	resData := map[string]interface{}{
		"signatureEncode64": signatureEncode64,
	}
	// Generate the response
	resp := &logical.Response{
		Data: resData,
	}

	return resp, nil
}

func (b *backend) handleWriteAccount(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	if req.ClientToken == "" {
		return nil, fmt.Errorf("client token empty")
	}

	// Check to make sure that kv pairs provided
	if len(req.Data) == 0 {
		return nil, fmt.Errorf("data must be provided to store in secret")
	}

	accountId := data.Get("accountId").(string)

	// JSON encode the data
	newAccount := Account{AccountId: accountId, PublicKey: "0x1234", PrivateKey: "0x9876"}
	b.Logger().Info(fmt.Sprintf("\n\n ClientToken=%v \n accountId=%v \n\n", req.ClientToken, accountId))

	buf, err := json.Marshal(newAccount)
	if err != nil {
		return nil, errwrap.Wrapf("json encoding failed: {{err}}", err)
	}

	// Store kv pairs in map at specified path
	entry := &logical.StorageEntry{
		Key:      req.ClientToken + "/" + accountId,
		Value:    buf,
		SealWrap: false,
	}
	if err = req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	resData := map[string]interface{}{
		"accountId": newAccount.AccountId,
		"status": true,
		"txHash": newAccount.PrivateKey,
	}
	// Generate the response
	resp := &logical.Response{
		Data: resData,
	}

	return resp, nil
}

func (b *backend) paths() []*framework.Path {
	return []*framework.Path{
		{
			Pattern: framework.MatchAllRegex("path"),

			Fields: map[string]*framework.FieldSchema{
				"path": {
					Type:        framework.TypeString,
					Description: "Specifies the path of the secret.",
				},
				"accountId": {
					Type:        framework.TypeString,
					Description: "Specifies the accountId of the secret.",
				},
			},

			Operations: map[logical.Operation]framework.OperationHandler{
				logical.ReadOperation: &framework.PathOperation{
					Callback: b.handleRead,
					Summary:  "Retrieve the secret from the map.",
				},
				logical.UpdateOperation: &framework.PathOperation{
					Callback: b.handleWrite,
					Summary:  "Store a secret at the specified location.",
				},
				logical.CreateOperation: &framework.PathOperation{
					Callback: b.handleWrite,
				},
				logical.DeleteOperation: &framework.PathOperation{
					Callback: b.handleDelete,
					Summary:  "Deletes the secret at the specified location.",
				},
			},

			ExistenceCheck: b.handleExistenceCheck,
		},
	}
}

func (b *backend) handleExistenceCheck(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	out, err := req.Storage.Get(ctx, req.Path)
	if err != nil {
		return false, errwrap.Wrapf("existence check failed: {{err}}", err)
	}

	return out != nil, nil
}

func (b *backend) handleRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	if req.ClientToken == "" {
		return nil, fmt.Errorf("client token empty")
	}

	path := data.Get("path").(string)

	// Decode the data
	var rawData Account
	entry, err := req.Storage.Get(ctx, req.ClientToken + "/" + path)
	if err != nil {
		return nil, err
	}
	fetchedData := entry.Value
	if fetchedData == nil {
		resp := logical.ErrorResponse("No value at %v%v", req.MountPoint, path)
		return resp, nil
	}

	if err := jsonutil.DecodeJSON(fetchedData, &rawData); err != nil {
		return nil, errwrap.Wrapf("json decoding failed: {{err}}", err)
	}


	b.Logger().Info(fmt.Sprintf("\n\n %v \n\n", rawData.PrivateKey))

	signatureEncode64 := rawData.PrivateKey + "-" + rawData.PublicKey
	resData := map[string]interface{}{
		"signatureEncode64": signatureEncode64,
	}
	// Generate the response
	resp := &logical.Response{
		Data: resData,
	}

	return resp, nil
}

func (b *backend) handleWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	if req.ClientToken == "" {
		return nil, fmt.Errorf("client token empty")
	}

	// Check to make sure that kv pairs provided
	if len(req.Data) == 0 {
		return nil, fmt.Errorf("data must be provided to store in secret")
	}

	path := data.Get("path").(string)
	accountId := data.Get("accountId").(string)
	// b.Logger().Warn(fmt.Sprintf("\n\n %v \n\n", path))
	// b.Logger().Info(fmt.Sprintf("\n\n %v \n\n", path))
	// b.Logger().Debug(fmt.Sprintf("\n\n %v \n\n", path))

	// JSON encode the data

	newAccount := Account{AccountId: accountId, PublicKey: "0x1234", PrivateKey: "0x9876"}
	b.Logger().Info(fmt.Sprintf("\n\n %v \n %v \n %v \n\n", newAccount.AccountId, newAccount.PublicKey, newAccount.PrivateKey))

	buf, err := json.Marshal(newAccount)
	if err != nil {
		return nil, errwrap.Wrapf("json encoding failed: {{err}}", err)
	}

	// Store kv pairs in map at specified path
	entry := &logical.StorageEntry{
		Key:      req.ClientToken + "/" + path,
		Value:    buf,
		SealWrap: false,
	}
	if err = req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	resData := map[string]interface{}{
		"accountId": newAccount.AccountId,
		"status": true,
		"txHash": newAccount.PrivateKey,
	}
	// Generate the response
	resp := &logical.Response{
		Data: resData,
	}

	return resp, nil
}

func (b *backend) handleDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	if req.ClientToken == "" {
		return nil, fmt.Errorf("client token empty")
	}

	path := data.Get("path").(string)

	// Remove entry for specified path
	if err := req.Storage.Delete(ctx, req.ClientToken + "/" + path); err != nil {
		return nil, err
	}

	return nil, nil
}

const mockHelp = `
The Mock backend is a dummy secrets backend that stores kv pairs in a map.
`
