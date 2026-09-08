package modem

import (
	"strings"

	wwanmodem "github.com/damonto/wwan-go/modem"
)

// SIMIdentity binds an observation to one activation of a SIM on a modem.
// Obtain it from Snapshot; its private version rejects delayed writes even
// when a SIM is removed and the same ICCID is subsequently reinserted.
type SIMIdentity struct {
	ICCID string
	Slot  uint32

	modem    *Modem
	revision uint64
}

// SetFallbackNumber records a network-derived number for the specified SIM.
// An empty number clears it. Modem-reported numbers always take precedence.
// It returns false if the SIM or modem activation has changed.
func (m *Modem) SetFallbackNumber(target SIMIdentity, number string) bool {
	if m == nil || target.ICCID == "" {
		return false
	}
	m.runtimeMu.Lock()
	defer m.runtimeMu.Unlock()
	if target != m.simIdentityLocked() {
		return false
	}
	m.fallbackNumber = strings.TrimSpace(number)
	return true
}

func (m *Modem) simIdentityLocked() SIMIdentity {
	identity := SIMIdentity{Slot: m.PrimarySIMSlot, modem: m, revision: m.simRevision}
	if m.SIM != nil && m.Status.SIM != wwanmodem.SIMStateAbsent {
		identity.ICCID = strings.TrimSpace(m.SIM.Identifier)
	}
	return identity
}

func (m *Modem) invalidateSIMIdentityLocked(previous SIMIdentity) {
	if previous == m.simIdentityLocked() {
		return
	}
	m.simRevision++
	m.fallbackNumber = ""
}
