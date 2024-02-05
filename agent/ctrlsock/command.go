package ctrlsock

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Command byte

func (c Command) String() string {
	switch c {
	case SetEnv:
		return "SetEnv"
	default:
		return fmt.Sprintf("Command(%d)", c)
	}
}

const (
	SetEnv Command = iota + 1
)

func readByte(r io.Reader) (byte, error) {
	var b byte
	if err := binary.Read(r, binary.BigEndian, &b); err != nil {
		return 0, err
	}
	return b, nil
}

func readString(r io.Reader) (string, error) {
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return "", err
	}
	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return "", err
	}
	return string(data), nil
}

func writeCommand(w io.Writer, c Command) error {
	return binary.Write(w, binary.BigEndian, c)
}

func writeString(w io.Writer, s string) error {
	if err := binary.Write(w, binary.BigEndian, uint32(len(s))); err != nil {
		return err
	}
	_, err := w.Write([]byte(s))
	return err
}

func readSetEnv(r io.Reader) (key string, value string, err error) {
	key, err = readString(r)
	if err != nil {
		return key, value, err
	}
	value, err = readString(r)
	if err != nil {
		return key, value, err
	}

	return key, value, nil
}

func writeSetEnv(w io.Writer, key string, value string) error {
	if err := writeCommand(w, SetEnv); err != nil {
		return err
	}
	if err := writeString(w, key); err != nil {
		return err
	}
	return writeString(w, value)
}
