package client

// 计算签名方式
type SignatureMethod string

const (
	HmacSHA256 SignatureMethod = "HmacSHA256"
	HmacSHA1   SignatureMethod = "HmacSHA1"
)

const signatureStr = `%s
%s

%s
%s
x-ca-key:%s
x-ca-nonce:%s
x-ca-signature-method:%s
x-ca-timestamp:%d
%s`
