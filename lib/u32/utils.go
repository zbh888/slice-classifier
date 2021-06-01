package u32

func BuildMatchGTP(gtpTEID uint32, ipv4 string, isRan bool) string {

	var outerHeader *IPV4Header
	outerHeader = &IPV4Header{
		Source:   ipv4,
		Protocol: PROTO_UDP,
		Set: &IPV4Fields{
			Protocol: true,
			Source:   true,
		},
	}
	var offset int = 16
	if isRan {
		offset = 8
		outerHeader = &IPV4Header{
			Destination: ipv4,
			Protocol:    PROTO_UDP,
			Set: &IPV4Fields{
				Protocol:    true,
				Destination: true,
			},
		}
	}

	protocols := []Protocol{
		outerHeader,
		&UDPHeader{
			SourcePort:      2152,
			DestinationPort: 2152,
			Set: &UDPFields{
				SourcePort:      true,
				DestinationPort: true,
			},
		},
		&GTPv1Header{
			HeaderOffset: offset,
			TEID:         gtpTEID,
			Set: &GTPv1Fields{
				TEID: true,
			},
		},
	}

	var m = NewU32(&protocols, 0)
	return m.Matches
}
