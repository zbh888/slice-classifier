// Package slicing holds the slice label informations
package slicing

import (
	"encoding/hex"
	"fmt"
)

// Label label representing the label : SLICE ID (8 bits) <=> PIPE (2 bits) <=> DSCP (6 bits) -------- 0xFF00 0x00C0 0x003F
type Label struct {
	SliceID uint8
	Pipe    bool
	DSCP    uint8
}

func (l *Label) Generate() string {

	var pipeu uint8

	if pipeu = (0 << 6); l.Pipe {
		pipeu = (1 << 6)
	}

	var label string = hex.EncodeToString([]byte{l.SliceID, pipeu + l.DSCP})
	return fmt.Sprintf("0x%s/0xFFFF", label)

}

func (l *Label) GenerateSliceID() string {
	var sliceID string = hex.EncodeToString([]byte{l.SliceID})
	return fmt.Sprintf("0x%s00/0xFF00", sliceID)
}

func (l *Label) GeneratePipe() string {
	var pipe string
	var pipeu uint8

	if pipeu = (0 << 6); l.Pipe {
		pipeu = (1 << 6)
	}

	pipe = hex.EncodeToString([]byte{pipeu})

	return fmt.Sprintf("0x00%s/0x00C0", pipe)
}

func (l *Label) GenerateDSCP() string {
	var dscp string = hex.EncodeToString([]byte{l.DSCP})
	return fmt.Sprintf("0x00%s/0x003F", dscp)
}

func (l *Label) GeneratePipeDSCP() string {
	var pipeu uint8

	if pipeu = (0 << 6); l.Pipe {
		pipeu = (1 << 6)
	}

	var pipeDscp string = hex.EncodeToString([]byte{pipeu + l.DSCP})
	return fmt.Sprintf("0x00%s/0x00FF", pipeDscp)
}

func (l *Label) GeneratePipeSliceID() string {
	var pipeu uint8

	if pipeu = (0 << 6); l.Pipe {
		pipeu = (1 << 6)
	}

	var pipeSliceID string = hex.EncodeToString([]byte{l.SliceID, pipeu})
	return fmt.Sprintf("0x%s/0xFFC0", pipeSliceID)
}

func NewLabel(slice_id uint8, pipe bool, dscp uint8) *Label {
	return &Label{SliceID: slice_id, Pipe: pipe, DSCP: dscp}
}

func Labels(slice_id uint8, dscp uint8) (pipe_label *Label, ipipe_label *Label) {
	pipe_label = NewLabel(slice_id, true, dscp)
	ipipe_label = NewLabel(slice_id, false, dscp)
	return
}
