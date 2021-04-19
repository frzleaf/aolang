package proxy

import (
	"bytes"
	"encoding/binary"
	"errors"
)

var InvalidMessageError = errors.New("invalid message")
var NotFoundConnectorError = errors.New("connector not found")

func ExtractMessageToConnectorIDAndData(data []byte) (int, []byte, error) {
	if len(data) < 6 {
		return -1, nil, InvalidMessageError
	}
	if bytes.Compare(data[0:len(MESSAGE_PREFIX_SIGN)], MESSAGE_PREFIX_SIGN) != 0 {
		return -1, nil, InvalidMessageError
	}
	connectorID := binary.BigEndian.Uint32(data[len(MESSAGE_PREFIX_SIGN) : len(MESSAGE_PREFIX_SIGN)+ConnectorIDLength])
	return int(connectorID), data[len(MESSAGE_PREFIX_SIGN)+ConnectorIDLength:], nil
}

func WrapMessage(connectorId int, data []byte) []byte {
	result := make([]byte, len(MESSAGE_PREFIX_SIGN)+ConnectorIDLength+len(data))
	copy(MESSAGE_PREFIX_SIGN, result[0:len(MESSAGE_PREFIX_SIGN)])
	binary.BigEndian.PutUint32(result[len(MESSAGE_PREFIX_SIGN):len(MESSAGE_PREFIX_SIGN)+ConnectorIDLength], uint32(connectorId))
	copy(data, result[len(MESSAGE_PREFIX_SIGN)+ConnectorIDLength:])
	return result
}
