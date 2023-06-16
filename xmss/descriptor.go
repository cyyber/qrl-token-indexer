package xmss

type SignatureType uint
type AddrFormatType uint
type HashFunction uint

const (
	DescriptorSize = 3
	ExtendedPKSize = 67
)

const (
	XMSSSig SignatureType = iota
)

const (
	SHA256_2X AddrFormatType = iota
)

const (
	SHA2_256 HashFunction = iota
	SHAKE_128
	SHAKE_256
)

type QRLDescriptor struct {
	hashFunction   HashFunction
	signatureType  SignatureType // Signature Type = XMSS = 0
	height         uint8
	addrFormatType AddrFormatType
}

func NewQRLDescriptor(height uint8, hashFunction HashFunction, signatureType SignatureType, addrFormatType AddrFormatType) *QRLDescriptor {
	return &QRLDescriptor{
		hashFunction:   hashFunction,
		signatureType:  signatureType,
		height:         height,
		addrFormatType: addrFormatType,
	}
}

func NewQRLDescriptorFromExtendedPK(extendedPK *[ExtendedPKSize]byte) *QRLDescriptor {
	return NewQRLDescriptorFromBytes(extendedPK[:DescriptorSize])
}

func NewQRLDescriptorFromBytes(descriptorBytes []uint8) *QRLDescriptor {
	if len(descriptorBytes) != 3 {
		panic("Descriptor size should be 3 bytes")
	}

	return &QRLDescriptor{
		hashFunction:   HashFunction(descriptorBytes[0] & 0x0f),
		signatureType:  SignatureType((descriptorBytes[0] >> 4) & 0x0f),
		height:         (descriptorBytes[1] & 0x0f) << 1,
		addrFormatType: AddrFormatType((descriptorBytes[1] & 0xF0) >> 4),
	}
}

func (d *QRLDescriptor) GetHeight() uint8 {
	return d.height
}

func (d *QRLDescriptor) GetHashFunction() HashFunction {
	return d.hashFunction
}

func (d *QRLDescriptor) GetSignatureType() SignatureType {
	return d.signatureType
}

func (d *QRLDescriptor) GetAddrFormatType() AddrFormatType {
	return d.addrFormatType
}

func (d *QRLDescriptor) GetBytes() [DescriptorSize]uint8 {
	var output [DescriptorSize]uint8
	output[0] = (uint8(d.signatureType) << 4) | (uint8(d.hashFunction) & 0x0F)
	output[1] = (uint8(d.addrFormatType) << 4) | ((d.height >> 1) & 0x0F)

	return output
}
