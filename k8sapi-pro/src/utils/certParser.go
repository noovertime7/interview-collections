package utils

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"k8sapi-pro/src/models"
)

func getCertType(alg x509.PublicKeyAlgorithm) string {
	switch alg {
	case x509.RSA:
		return "RSA"
	case x509.DSA:
		return "DSA"
	case x509.ECDSA:
		return "ECDSA"
	case x509.UnknownPublicKeyAlgorithm:
		return "Unknow"
	}
	return "Unknow"
}

// 解析证书
func ParseCert(crt []byte) *models.CertModel {
	var cert tls.Certificate
	//解码证书
	certBlock, restPEMBlock := pem.Decode(crt)
	if certBlock == nil {
		return nil
	}
	cert.Certificate = append(cert.Certificate, certBlock.Bytes)
	//处理证书链
	certBlockChain, _ := pem.Decode(restPEMBlock)
	if certBlockChain != nil {
		cert.Certificate = append(cert.Certificate, certBlockChain.Bytes)
	}

	//解析证书
	x509Cert, err := x509.ParseCertificate(certBlock.Bytes)

	if err != nil {
		return nil
	} else {
		return &models.CertModel{
			CN:        x509Cert.Subject.CommonName,
			Issuer:    x509Cert.Issuer.CommonName,
			Algorithm: getCertType(x509Cert.PublicKeyAlgorithm),
			BeginTime: x509Cert.NotBefore.Format("2006-01-02 15:03:04"),
			EndTime:   x509Cert.NotAfter.Format("2006-01-02 15:03:04"),
		}
	}
}
