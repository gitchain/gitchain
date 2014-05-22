package api

import (
	"net/http"

	"github.com/gitchain/gitchain/keys"
	"github.com/gitchain/gitchain/server"
	"github.com/inconshreveable/log15"
)

// KeySerice
type KeyService struct {
	srv *server.T
	log log15.Logger
}

type GeneratePrivateKeyArgs struct {
	Alias string
}

type GeneratePrivateKeyReply struct {
	Success   bool
	PublicKey string
}

func (service *KeyService) GeneratePrivateKey(r *http.Request, args *GeneratePrivateKeyArgs, reply *GeneratePrivateKeyReply) error {
	key, err := keys.GenerateECDSA()
	if err != nil {
		reply.Success = false
		return err
	}
	err = service.srv.DB.PutKey(args.Alias, key, false)
	if err != nil {
		reply.Success = false
		return err
	}
	reply.PublicKey = keys.ECDSAPublicKeyToString(key.PublicKey)
	reply.Success = true
	return nil
}

type ListPrivateKeysArgs struct {
}

type ListPrivateKeysReply struct {
	Aliases []string
}

func (service *KeyService) ListPrivateKeys(r *http.Request, args *ListPrivateKeysArgs, reply *ListPrivateKeysReply) error {
	reply.Aliases = service.srv.DB.ListKeys()
	return nil
}

type SetMainKeyArgs struct {
	Alias string
}

type SetMainKeyReply struct {
	Success bool
}

func (service *KeyService) SetMainKey(r *http.Request, args *SetMainKeyArgs, reply *SetMainKeyReply) error {
	key, err := service.srv.DB.GetKey(args.Alias)
	if err != nil {
		return err
	}
	if key != nil {
		err := service.srv.DB.PutKey(args.Alias, key, true)
		if err != nil {
			return err
		}
		reply.Success = true
	} else {
		reply.Success = false
	}
	return nil
}

type GetMainKeyArgs struct {
}

type GetMainKeyReply struct {
	Alias string
}

func (service *KeyService) GetMainKey(r *http.Request, args *GetMainKeyArgs, reply *GetMainKeyReply) error {
	allKeys := service.srv.DB.ListKeys()
	mainKey, err := service.srv.DB.GetMainKey()
	if err != nil {
		return err
	}
	for i := range allKeys {
		key, err := service.srv.DB.GetKey(allKeys[i])
		if err != nil {
			return err
		}
		if equal, _ := keys.EqualECDSAPrivateKeys(mainKey, key); equal {
			reply.Alias = allKeys[i]
		}
	}
	return nil
}
