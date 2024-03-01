package core

type Link struct {
	ObjectId string `json:"objectId"`
	Size     int64  `json:"size"`
}

type VaultLink struct {
	VaultId string `json:"vaultId"`
	KeyId   string `json:"keyId"`
}

func NewVaultLink(vaultId string, keyId string) VaultLink {
	return VaultLink{
		VaultId: vaultId,
		KeyId:   keyId,
	}
}
