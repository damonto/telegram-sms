package modemmanager

/* Enums */

// MMModemCapability Flags describing one or more of the general access technology families that a modem supports.
type MMModemCapability uint32

//go:generate stringer -type=MMModemCapability -trimprefix=MmModemCapability
const (
	MmModemCapabilityNone        MMModemCapability = 0          // Modem has no capabilities.
	MmModemCapabilityPots        MMModemCapability = 1 << 0     // Modem supports the analog wired telephone network (ie 56k dialup) and does not have wireless/cellular capabilities.
	MmModemCapabilityCdmaEvdo    MMModemCapability = 1 << 1     // Modem supports at least one of CDMA 1xRTT, EVDO revision 0, EVDO revision A, or EVDO revision B.
	MmModemCapabilityGsmUmts     MMModemCapability = 1 << 2     // Modem supports at least one of GSM, GPRS, EDGE, UMTS, HSDPA, HSUPA, or HSPA+ packet switched data capability.
	MmModemCapabilityLte         MMModemCapability = 1 << 3     // Modem has LTE data capability.
	MmModemCapabilityLteAdvanced MMModemCapability = 1 << 4     // Modem has LTE Advanced data capability.
	MmModemCapabilityIridium     MMModemCapability = 1 << 5     //Modem has Iridium capabilities.
	MmModemCapabilityAny         MMModemCapability = 0xFFFFFFFF // Mask specifying all capabilities.
)

// GetAllCapabilities returns all capabilities
func (c MMModemCapability) GetAllCapabilities() []MMModemCapability {
	return []MMModemCapability{MmModemCapabilityPots, MmModemCapabilityCdmaEvdo, MmModemCapabilityGsmUmts, MmModemCapabilityLte,
		MmModemCapabilityLteAdvanced, MmModemCapabilityIridium,
	}
}

// BitmaskToSlice bitmask to slice
func (c MMModemCapability) BitmaskToSlice(bitmask uint32) (capabilities []MMModemCapability) {
	if bitmask == 0 {
		return
	}
	for idx, x := range c.GetAllCapabilities() {
		if bitmask&(1<<idx) > 0 {
			capabilities = append(capabilities, x)
		}
	}
	return
}

// SliceToBitmask slice to bitmask
func (c MMModemCapability) SliceToBitmask(capabilities []MMModemCapability) (bitmask uint32) {
	bitmask = 0
	for idx, x := range c.GetAllCapabilities() {
		for _, y := range capabilities {
			if x == y {
				bitmask = bitmask | (1 << idx)
			}
		}
	}
	return bitmask
}

// MMModemLock Possible lock reasons.
type MMModemLock uint32

//go:generate stringer -type=MMModemLock -trimprefix=MmModemLock
const (
	MmModemLockUnknown     MMModemLock = 0  // Lock reason unknown.
	MmModemLockNone        MMModemLock = 1  // Modem is unlocked.
	MmModemLockSimPin      MMModemLock = 2  // SIM requires the PIN code.
	MmModemLockSimPin2     MMModemLock = 3  // SIM requires the PIN2 code.
	MmModemLockSimPuk      MMModemLock = 4  // SIM requires the PUK code.
	MmModemLockSimPuk2     MMModemLock = 5  // SIM requires the PUK2 code.
	MmModemLockPhSpPin     MMModemLock = 6  // Modem requires the service provider PIN code.
	MmModemLockPhSpPuk     MMModemLock = 7  // Modem requires the service provider PUK code.
	MmModemLockPhNetPin    MMModemLock = 8  // Modem requires the network PIN code.
	MmModemLockPhNetPuk    MMModemLock = 9  // Modem requires the network PUK code.
	MmModemLockPhSimPin    MMModemLock = 10 // Modem requires the PIN code.
	MmModemLockPhCorpPin   MMModemLock = 11 // Modem requires the corporate PIN code.
	MmModemLockPhCorpPuk   MMModemLock = 12 // Modem requires the corporate PUK code.
	MmModemLockPhFsimPin   MMModemLock = 13 // Modem requires the PH-FSIM PIN code.
	MmModemLockPhFsimPuk   MMModemLock = 14 // Modem requires the PH-FSIM PUK code.
	MmModemLockPhNetsubPin MMModemLock = 15 // Modem requires the network subset PIN code.
	MmModemLockPhNetsubPuk MMModemLock = 16 // Modem requires the network subset PUK code.
)

// MMModemState Possible modem states.
type MMModemState int32

//go:generate stringer -type=MMModemState -trimprefix=MmModemState
const (
	MmModemStateFailed        MMModemState = -1 // The modem is unusable.
	MmModemStateUnknown       MMModemState = 0  // State unknown or not reportable.
	MmModemStateInitializing  MMModemState = 1  // The modem is currently being initialized.
	MmModemStateLocked        MMModemState = 2  // The modem needs to be unlocked.
	MmModemStateDisabled      MMModemState = 3  // The modem is not enabled and is powered down.
	MmModemStateDisabling     MMModemState = 4  // The modem is currently transitioning to the @MmModemStateDisabled state.
	MmModemStateEnabling      MMModemState = 5  // The modem is currently transitioning to the @MmModemStateEnabled state.
	MmModemStateEnabled       MMModemState = 6  // The modem is enabled and powered on but not registered with a network provider and not available for data connections.
	MmModemStateSearching     MMModemState = 7  // The modem is searching for a network provider to register with.
	MmModemStateRegistered    MMModemState = 8  // The modem is registered with a network provider, and data connections and messaging may be available for use.
	MmModemStateDisconnecting MMModemState = 9  // The modem is disconnecting and deactivating the last active packet data bearer. This state will not be entered if more than one packet data bearer is active and one of the active bearers is deactivated.
	MmModemStateConnecting    MMModemState = 10 // The modem is activating and connecting the first packet data bearer. Subsequent bearer activations when another bearer is already active do not cause this state to be entered.
	MmModemStateConnected     MMModemState = 11 // One or more packet data bearers is active and connected.

)

// MMModemStateFailedReason Power state of the modem.
type MMModemStateFailedReason uint32

//go:generate stringer -type=MMModemStateFailedReason -trimprefix=MmModemStateFailedReason
const (
	MmModemStateFailedReasonNone       MMModemStateFailedReason = 0 // No error.
	MmModemStateFailedReasonUnknown    MMModemStateFailedReason = 1 // Unknown error.
	MmModemStateFailedReasonSimMissing MMModemStateFailedReason = 2 // SIM is required but missing.
	MmModemStateFailedReasonSimError   MMModemStateFailedReason = 3 // SIM is available, but unusable (e.g. permanently locked).

)

// MMModemPowerState Power state of the modem.
type MMModemPowerState uint32

//go:generate stringer -type=MMModemPowerState -trimprefix=MmModemPowerState
const (
	MmModemPowerStateUnknown MMModemPowerState = 0 // Unknown power state.
	MmModemPowerStateOff     MMModemPowerState = 1 // Off.
	MmModemPowerStateLow     MMModemPowerState = 2 // Low-power mode.
	MmModemPowerStateOn      MMModemPowerState = 3 // Full power mode.

)

// MMModemStateChangeReason Possible reasons to have changed the modem state.
type MMModemStateChangeReason uint32

//go:generate stringer -type=MMModemStateChangeReason -trimprefix=MmModemStateChangeReason
const (
	MmModemStateChangeReasonUnknown       MMModemStateChangeReason = 0 // Reason unknown or not reportable.
	MmModemStateChangeReasonUserRequested MMModemStateChangeReason = 1 // State change was requested by an interface user.
	MmModemStateChangeReasonSuspend       MMModemStateChangeReason = 2 // State change was caused by a system suspend.
	MmModemStateChangeReasonFailure       MMModemStateChangeReason = 3 // State change was caused by an unrecoverable error.

)

// MMModemAccessTechnology Describes various access technologies that a device uses when registered with or connected to a network.
type MMModemAccessTechnology uint32

//go:generate stringer -type=MMModemAccessTechnology  -trimprefix=MmModemAccessTechnology
const (
	MmModemAccessTechnologyUnknown    MMModemAccessTechnology = 0          // The access technology used is unknown.
	MmModemAccessTechnologyPots       MMModemAccessTechnology = 1 << 0     // Analog wireline telephone.
	MmModemAccessTechnologyGsm        MMModemAccessTechnology = 1 << 1     // GSM.
	MmModemAccessTechnologyGsmCompact MMModemAccessTechnology = 1 << 2     // Compact GSM.
	MmModemAccessTechnologyGprs       MMModemAccessTechnology = 1 << 3     // GPRS.
	MmModemAccessTechnologyEdge       MMModemAccessTechnology = 1 << 4     // EDGE (ETSI 27.007: "GSM w/EGPRS").
	MmModemAccessTechnologyUmts       MMModemAccessTechnology = 1 << 5     // UMTS (ETSI 27.007: "UTRAN").
	MmModemAccessTechnologyHsdpa      MMModemAccessTechnology = 1 << 6     // HSDPA (ETSI 27.007: "UTRAN w/HSDPA").
	MmModemAccessTechnologyHsupa      MMModemAccessTechnology = 1 << 7     // HSUPA (ETSI 27.007: "UTRAN w/HSUPA").
	MmModemAccessTechnologyHspa       MMModemAccessTechnology = 1 << 8     // HSPA (ETSI 27.007: "UTRAN w/HSDPA and HSUPA").
	MmModemAccessTechnologyHspaPlus   MMModemAccessTechnology = 1 << 9     // HSPA+ (ETSI 27.007: "UTRAN w/HSPA+").
	MmModemAccessTechnology1xrtt      MMModemAccessTechnology = 1 << 10    // CDMA2000 1xRTT.
	MmModemAccessTechnologyEvdo0      MMModemAccessTechnology = 1 << 11    // CDMA2000 EVDO revision 0.
	MmModemAccessTechnologyEvdoa      MMModemAccessTechnology = 1 << 12    // CDMA2000 EVDO revision A.
	MmModemAccessTechnologyEvdob      MMModemAccessTechnology = 1 << 13    // CDMA2000 EVDO revision B.
	MmModemAccessTechnologyLte        MMModemAccessTechnology = 1 << 14    // LTE (ETSI 27.007: "E-UTRAN")
	MmModemAccessTechnologyAny        MMModemAccessTechnology = 0xFFFFFFFF // Mask specifying all access technologies.
)

// GetAllTechnologies returns all technologies
func (t MMModemAccessTechnology) GetAllTechnologies() []MMModemAccessTechnology {
	var technologies = []MMModemAccessTechnology{MmModemAccessTechnologyPots, MmModemAccessTechnologyGsm, MmModemAccessTechnologyGsmCompact,
		MmModemAccessTechnologyGprs, MmModemAccessTechnologyEdge, MmModemAccessTechnologyUmts, MmModemAccessTechnologyHsdpa, MmModemAccessTechnologyHsupa, MmModemAccessTechnologyHspa,
		MmModemAccessTechnologyHspaPlus, MmModemAccessTechnology1xrtt, MmModemAccessTechnologyEvdo0, MmModemAccessTechnologyEvdoa, MmModemAccessTechnologyEvdob, MmModemAccessTechnologyLte,
	}
	return technologies
}

// BitmaskToSlice bitmask to slice
func (t MMModemAccessTechnology) BitmaskToSlice(bitmask uint32) (technologies []MMModemAccessTechnology) {
	if bitmask == 0 {
		return
	}
	for idx, x := range t.GetAllTechnologies() {
		if bitmask&(1<<idx) > 0 {
			technologies = append(technologies, x)
		}
	}
	return technologies
}

// SliceToBitmask slice to bitmask
func (t MMModemAccessTechnology) SliceToBitmask(technologies []MMModemAccessTechnology) (bitmask uint32) {
	bitmask = 0
	for idx, x := range t.GetAllTechnologies() {
		for _, y := range technologies {
			if x == y {
				bitmask = bitmask | (1 << idx)
			}
		}
	}
	return bitmask
}

// MMModemMode Bitfield to indicate which access modes are supported, allowed or preferred in a given device.
type MMModemMode uint32

//go:generate stringer -type=MMModemMode -trimprefix=MmModemMode
const (
	MmModemModeNone MMModemMode = 0          // None
	MmModemModeCs   MMModemMode = 1 << 0     // CSD, GSM, and other circuit-switched technologies.
	MmModemMode2g   MMModemMode = 1 << 1     // GPRS, EDGE.
	MmModemMode3g   MMModemMode = 1 << 2     // UMTS, HSxPA.
	MmModemMode4g   MMModemMode = 1 << 3     // LTE.
	MmModemModeAny  MMModemMode = 0xFFFFFFFF // Any mode can be used (only this value allowed for POTS modems).

)

// GetAllModes returns all modes
func (m MMModemMode) GetAllModes() []MMModemMode {
	return []MMModemMode{MmModemModeCs, MmModemMode2g, MmModemMode3g, MmModemMode4g}
}

// BitmaskToSlice bitmask to slice
func (m MMModemMode) BitmaskToSlice(bitmask uint32) (modes []MMModemMode) {
	if bitmask == 0 {
		return
	}
	for idx, x := range m.GetAllModes() {
		if bitmask&(1<<idx) > 0 {
			modes = append(modes, x)
		}
	}
	return
}

// SliceToBitmask slice to bitmask
func (m MMModemMode) SliceToBitmask(modes []MMModemMode) (bitmask uint32) {
	bitmask = 0
	for idx, x := range m.GetAllModes() {
		for _, y := range modes {
			if x == y {
				bitmask = bitmask | (1 << idx)
			}
		}
	}
	return bitmask
}

// MMModemBand Radio bands supported by the device when connecting to a mobile network.
type MMModemBand uint32

//go:generate stringer -type=MMModemBand -trimprefix=MmModemBand
const (
	MmModemBandUnknown MMModemBand = 0 // Unknown or invalid band.
	/* GSM/UMTS bands */
	MmModemBandEgsm   MMModemBand = 1  // GSM/GPRS/EDGE 900 MHz.
	MmModemBandDcs    MMModemBand = 2  // GSM/GPRS/EDGE 1800 MHz.
	MmModemBandPcs    MMModemBand = 3  // GSM/GPRS/EDGE 1900 MHz.
	MmModemBandG850   MMModemBand = 4  // GSM/GPRS/EDGE 850 MHz.
	MmModemBandUtran1 MMModemBand = 5  // UMTS 2100 MHz (IMT, UTRAN band 1).
	MmModemBandUtran3 MMModemBand = 6  // UMTS 1800 MHz (DCS, UTRAN band 3).
	MmModemBandUtran4 MMModemBand = 7  // UMTS 1700 MHz (AWS A-F, UTRAN band 4).
	MmModemBandUtran6 MMModemBand = 8  // UMTS 800 MHz (UTRAN band 6).
	MmModemBandUtran5 MMModemBand = 9  // UMTS 850 MHz (CLR, UTRAN band 5).
	MmModemBandUtran8 MMModemBand = 10 //UMTS 900 MHz (E-GSM, UTRAN band 8).
	MmModemBandUtran9 MMModemBand = 11 // UMTS 1700 MHz (UTRAN band 9).
	MmModemBandUtran2 MMModemBand = 12 //  UMTS 1900 MHz (PCS A-F, UTRAN band 2).
	MmModemBandUtran7 MMModemBand = 13 // UMTS 2600 MHz (IMT-E, UTRAN band 7).
	MmModemBandG450   MMModemBand = 14 // GSM/GPRS/EDGE 450 MHz.
	MmModemBandG480   MMModemBand = 15 // GSM/GPRS/EDGE 480 MHz.
	MmModemBandG750   MMModemBand = 16 // GSM/GPRS/EDGE 750 MHz.
	MmModemBandG380   MMModemBand = 17 // GSM/GPRS/EDGE 380 MHz.
	MmModemBandG410   MMModemBand = 18 // GSM/GPRS/EDGE 410 MHz.
	MmModemBandG710   MMModemBand = 19 // GSM/GPRS/EDGE 710 MHz.
	MmModemBandG810   MMModemBand = 20 // GSM/GPRS/EDGE 810 MHz.
	/* LTE bands */
	MmModemBandEutran1  MMModemBand = 31  // E-UTRAN band 1.
	MmModemBandEutran2  MMModemBand = 32  // E-UTRAN band 2.
	MmModemBandEutran3  MMModemBand = 33  // E-UTRAN band 3.
	MmModemBandEutran4  MMModemBand = 34  // E-UTRAN band 4.
	MmModemBandEutran5  MMModemBand = 35  // E-UTRAN band 5.
	MmModemBandEutran6  MMModemBand = 36  // E-UTRAN band 6.
	MmModemBandEutran7  MMModemBand = 37  // E-UTRAN band 7.
	MmModemBandEutran8  MMModemBand = 38  // E-UTRAN band 8.
	MmModemBandEutran9  MMModemBand = 39  // E-UTRAN band 9.
	MmModemBandEutran10 MMModemBand = 40  // E-UTRAN band 10.
	MmModemBandEutran11 MMModemBand = 41  // E-UTRAN band 11.
	MmModemBandEutran12 MMModemBand = 42  // E-UTRAN band 12.
	MmModemBandEutran13 MMModemBand = 43  // E-UTRAN band 13.
	MmModemBandEutran14 MMModemBand = 44  // E-UTRAN band 14.
	MmModemBandEutran17 MMModemBand = 47  // E-UTRAN band 17.
	MmModemBandEutran18 MMModemBand = 48  // E-UTRAN band 18.
	MmModemBandEutran19 MMModemBand = 49  // E-UTRAN band 19.
	MmModemBandEutran20 MMModemBand = 50  // E-UTRAN band 20.
	MmModemBandEutran21 MMModemBand = 51  // E-UTRAN band 21.
	MmModemBandEutran22 MMModemBand = 52  // E-UTRAN band 22.
	MmModemBandEutran23 MMModemBand = 53  // E-UTRAN band 23.
	MmModemBandEutran24 MMModemBand = 54  // E-UTRAN band 24.
	MmModemBandEutran25 MMModemBand = 55  // E-UTRAN band 25.
	MmModemBandEutran26 MMModemBand = 56  // E-UTRAN band 26.
	MmModemBandEutran27 MMModemBand = 57  // E-UTRAN band 27.
	MmModemBandEutran28 MMModemBand = 58  // E-UTRAN band 28.
	MmModemBandEutran29 MMModemBand = 59  // E-UTRAN band 29.
	MmModemBandEutran30 MMModemBand = 60  // E-UTRAN band 30.
	MmModemBandEutran31 MMModemBand = 61  // E-UTRAN band 31.
	MmModemBandEutran32 MMModemBand = 62  // E-UTRAN band 32.
	MmModemBandEutran33 MMModemBand = 63  // E-UTRAN band 33.
	MmModemBandEutran34 MMModemBand = 64  // E-UTRAN band 34.
	MmModemBandEutran35 MMModemBand = 65  // E-UTRAN band 35.
	MmModemBandEutran36 MMModemBand = 66  // E-UTRAN band 36.
	MmModemBandEutran37 MMModemBand = 67  // E-UTRAN band 37.
	MmModemBandEutran38 MMModemBand = 68  // E-UTRAN band 38.
	MmModemBandEutran39 MMModemBand = 69  // E-UTRAN band 39.
	MmModemBandEutran40 MMModemBand = 70  // E-UTRAN band 40
	MmModemBandEutran41 MMModemBand = 71  // E-UTRAN band 41.
	MmModemBandEutran42 MMModemBand = 72  // E-UTRAN band 42.
	MmModemBandEutran43 MMModemBand = 73  // E-UTRAN band 43.
	MmModemBandEutran44 MMModemBand = 74  // E-UTRAN band 44.
	MmModemBandEutran45 MMModemBand = 75  // E-UTRAN band 45.
	MmModemBandEutran46 MMModemBand = 76  // E-UTRAN band 46.
	MmModemBandEutran47 MMModemBand = 77  // E-UTRAN band 47.
	MmModemBandEutran48 MMModemBand = 78  // E-UTRAN band 48.
	MmModemBandEutran49 MMModemBand = 79  // E-UTRAN band 49.
	MmModemBandEutran50 MMModemBand = 80  // E-UTRAN band 50.
	MmModemBandEutran51 MMModemBand = 81  // E-UTRAN band 51.
	MmModemBandEutran52 MMModemBand = 82  // E-UTRAN band 52.
	MmModemBandEutran53 MMModemBand = 83  // E-UTRAN band 53.
	MmModemBandEutran54 MMModemBand = 84  // E-UTRAN band 54.
	MmModemBandEutran55 MMModemBand = 85  // E-UTRAN band 55.
	MmModemBandEutran56 MMModemBand = 86  // E-UTRAN band 56.
	MmModemBandEutran57 MMModemBand = 87  // E-UTRAN band 57.
	MmModemBandEutran58 MMModemBand = 88  // E-UTRAN band 58.
	MmModemBandEutran59 MMModemBand = 89  // E-UTRAN band 59.
	MmModemBandEutran60 MMModemBand = 90  // E-UTRAN band 60.
	MmModemBandEutran61 MMModemBand = 91  // E-UTRAN band 61.
	MmModemBandEutran62 MMModemBand = 92  // E-UTRAN band 62.
	MmModemBandEutran63 MMModemBand = 93  // E-UTRAN band 63.
	MmModemBandEutran64 MMModemBand = 94  // E-UTRAN band 64.
	MmModemBandEutran65 MMModemBand = 95  // E-UTRAN band 65.
	MmModemBandEutran66 MMModemBand = 96  // E-UTRAN band 66.
	MmModemBandEutran67 MMModemBand = 97  // E-UTRAN band 67.
	MmModemBandEutran68 MMModemBand = 98  // E-UTRAN band 68.
	MmModemBandEutran69 MMModemBand = 99  // E-UTRAN band 69.
	MmModemBandEutran70 MMModemBand = 100 // E-UTRAN band 70.
	MmModemBandEutran71 MMModemBand = 101 // E-UTRAN band 71.
	/* CDMA Band Classes (see 3GPP2 C.S0057-C) */
	MmModemBandCdmaBc0  MMModemBand = 128 // CDMA Band Class 0 (US Cellular 850MHz).
	MmModemBandCdmaBc1  MMModemBand = 129 // CDMA Band Class 1 (US PCS 1900MHz).
	MmModemBandCdmaBc2  MMModemBand = 130 // CDMA Band Class 2 (UK TACS 900MHz).
	MmModemBandCdmaBc3  MMModemBand = 131 // CDMA Band Class 3 (Japanese TACS).
	MmModemBandCdmaBc4  MMModemBand = 132 // CDMA Band Class 4 (Korean PCS).
	MmModemBandCdmaBc5  MMModemBand = 134 // CDMA Band Class 5 (NMT 450MHz).
	MmModemBandCdmaBc6  MMModemBand = 135 // CDMA Band Class 6 (IMT2000 2100MHz).
	MmModemBandCdmaBc7  MMModemBand = 136 // CDMA Band Class 7 (Cellular 700MHz).
	MmModemBandCdmaBc8  MMModemBand = 137 // CDMA Band Class 8 (1800MHz).
	MmModemBandCdmaBc9  MMModemBand = 138 // CDMA Band Class 9 (900MHz).
	MmModemBandCdmaBc10 MMModemBand = 139 // (US Secondary 800).
	MmModemBandCdmaBc11 MMModemBand = 140 // (European PAMR 400MHz).
	MmModemBandCdmaBc12 MMModemBand = 141 // (PAMR 800MHz).
	MmModemBandCdmaBc13 MMModemBand = 142 // (IMT2000 2500MHz Expansion).
	MmModemBandCdmaBc14 MMModemBand = 143 // (More US PCS 1900MHz).
	MmModemBandCdmaBc15 MMModemBand = 144 // (AWS 1700MHz).
	MmModemBandCdmaBc16 MMModemBand = 145 // (US 2500MHz).
	MmModemBandCdmaBc17 MMModemBand = 146 // (US 2500MHz Forward Link Only).
	MmModemBandCdmaBc18 MMModemBand = 147 // (US 700MHz Public Safety).
	MmModemBandCdmaBc19 MMModemBand = 148 // (US Lower 700MHz).
	/* Additional UMTS bands:
	 *  15-18 reserved
	 *  23-24 reserved
	 *  27-31 reserved
	 */
	MmModemBandUtran10 MMModemBand = 210 // UMTS 1700 MHz (EAWS A-G, UTRAN band 10).
	MmModemBandUtran11 MMModemBand = 211 // UMTS 1500 MHz (LPDC, UTRAN band 11).
	MmModemBandUtran12 MMModemBand = 212 // UMTS 700 MHz (LSMH A/B/C, UTRAN band 12)
	MmModemBandUtran13 MMModemBand = 213 // UMTS 700 MHz (USMH C, UTRAN band 13)
	MmModemBandUtran14 MMModemBand = 214 // UMTS 700 MHz (USMH D, UTRAN band 14)
	MmModemBandUtran19 MMModemBand = 219 // UMTS 800 MHz (UTRAN band 19).
	MmModemBandUtran20 MMModemBand = 220 // UMTS 800 MHz (EUDD, UTRAN band 20).
	MmModemBandUtran21 MMModemBand = 221 // UMTS 1500 MHz (UPDC, UTRAN band 21).
	MmModemBandUtran22 MMModemBand = 222 // UMTS 3500 MHz (UTRAN band 22).
	MmModemBandUtran25 MMModemBand = 225 // UMTS 1900 MHz (EPCS A-G, UTRAN band 25).
	MmModemBandUtran26 MMModemBand = 226 // UMTS 850 MHz (ECLR, UTRAN band 26).
	MmModemBandUtran32 MMModemBand = 232 // UMTS 1500 MHz (L-band, UTRAN band 32).
	/* All/Any */
	MmModemBandAny MMModemBand = 256 // For certain operations, allow the modem to select a band automatically.
)

// MMModemPortType Type of modem port.
type MMModemPortType uint32

//go:generate stringer -type=MMModemPortType -trimprefix=MmModemPortType
const (
	MmModemPortTypeUnknown MMModemPortType = 1 // Unknown.
	MmModemPortTypeNet     MMModemPortType = 2 // Net port.
	MmModemPortTypeAt      MMModemPortType = 3 // AT port.
	MmModemPortTypeQcdm    MMModemPortType = 4 // QCDM port.
	MmModemPortTypeGps     MMModemPortType = 5 // GPS port.
	MmModemPortTypeQmi     MMModemPortType = 6 // QMI port.
	MmModemPortTypeMbim    MMModemPortType = 7 // MBIM port.
	MmModemPortTypeAudio   MMModemPortType = 8 // Audio port.

)

// MMSmsPduType Type of PDUs used in the SMS.
type MMSmsPduType uint32

//go:generate stringer -type=MMSmsPduType -trimprefix=MmSmsPduType
const (
	MmSmsPduTypeUnknown                     MMSmsPduType = 0  // Unknown type.
	MmSmsPduTypeDeliver                     MMSmsPduType = 1  // 3GPP Mobile-Terminated (MT) message.
	MmSmsPduTypeSubmit                      MMSmsPduType = 2  // 3GPP Mobile-Originated (MO) message.
	MmSmsPduTypeStatusReport                MMSmsPduType = 3  // 3GPP status report (MT).
	MmSmsPduTypeCdmaDeliver                 MMSmsPduType = 32 // 3GPP2 Mobile-Terminated (MT) message.
	MmSmsPduTypeCdmaSubmit                  MMSmsPduType = 33 // 3GPP2 Mobile-Originated (MO) message.
	MmSmsPduTypeCdmaCancellation            MMSmsPduType = 34 // 3GPP2 Cancellation (MO) message.
	MmSmsPduTypeCdmaDeliveryAcknowledgement MMSmsPduType = 35 // 3GPP2 Delivery Acknowledgement (MT) message.
	MmSmsPduTypeCdmaUserAcknowledgement     MMSmsPduType = 36 // 3GPP2 User Acknowledgement (MT or MO) message.
	MmSmsPduTypeCdmaReadAcknowledgement     MMSmsPduType = 37 // 3GPP2 Read Acknowledgement (MT or MO) message.

)

// MMSmsState State of a given SMS.
type MMSmsState uint32

//go:generate stringer -type=MMSmsState -trimprefix=MmSmsState
const (
	MmSmsStateUnknown   MMSmsState = 0 // State unknown or not reportable.
	MmSmsStateStored    MMSmsState = 1 // The message has been neither received nor yet sent.
	MmSmsStateReceiving MMSmsState = 2 // The message is being received but is not yet complete.
	MmSmsStateReceived  MMSmsState = 3 // The message has been completely received.
	MmSmsStateSending   MMSmsState = 4 // The message is queued for delivery.
	MmSmsStateSent      MMSmsState = 5 // The message was successfully sent.

)

// MMSmsDeliveryState Known SMS delivery states as defined in 3GPP TS 03.40 and  3GPP2 N.S0005-O, section 6.5.2.125. States out of the known ranges may also be valid (either reserved or SC-specific).
type MMSmsDeliveryState uint32

//go:generate stringer -type=MMSmsDeliveryState -trimprefix=MmSmsDeliveryState
const (
	/* Completed deliveries */
	MmSmsDeliveryStateCompletedReceived             MMSmsDeliveryState = 0x00 // Delivery completed, message received by the SME.
	MmSmsDeliveryStateCompletedForwardedUnconfirmed MMSmsDeliveryState = 0x01 // Forwarded by the SC to the SME but the SC is unable to confirm delivery.
	MmSmsDeliveryStateCompletedReplacedBySc         MMSmsDeliveryState = 0x02 // Message replaced by the SC.

	/* Temporary failures */
	MmSmsDeliveryStateTemporaryErrorCongestion        MMSmsDeliveryState = 0x20 // Temporary error, congestion.
	MmSmsDeliveryStateTemporaryErrorSmeBusy           MMSmsDeliveryState = 0x21 // Temporary error, SME busy.
	MmSmsDeliveryStateTemporaryErrorNoResponseFromSme MMSmsDeliveryState = 0x22 // Temporary error, no response from the SME.
	MmSmsDeliveryStateTemporaryErrorServiceRejected   MMSmsDeliveryState = 0x23 // Temporary error, service rejected.
	MmSmsDeliveryStateTemporaryErrorQosNotAvailable   MMSmsDeliveryState = 0x24 // Temporary error, QoS not available.
	MmSmsDeliveryStateTemporaryErrorInSme             MMSmsDeliveryState = 0x25 // Temporary error in the SME.

	/* Permanent failures */
	MmSmsDeliveryStateErrorRemoteProcedure           MMSmsDeliveryState = 0x40 // Permanent remote procedure error.
	MmSmsDeliveryStateErrorIncompatibleDestination   MMSmsDeliveryState = 0x41 // Permanent error, incompatible destination.
	MmSmsDeliveryStateErrorConnectionRejected        MMSmsDeliveryState = 0x42 // Permanent error, connection rejected by the SME.
	MmSmsDeliveryStateErrorNotObtainable             MMSmsDeliveryState = 0x43 // Permanent error, not obtainable.
	MmSmsDeliveryStateErrorQosNotAvailable           MMSmsDeliveryState = 0x44 // Permanent error, QoS not available.
	MmSmsDeliveryStateErrorNoInterworkingAvailable   MMSmsDeliveryState = 0x45 // Permanent error, no interworking available.
	MmSmsDeliveryStateErrorValidityPeriodExpired     MMSmsDeliveryState = 0x46 // Permanent error, message validity period expired.
	MmSmsDeliveryStateErrorDeletedByOriginatingSme   MMSmsDeliveryState = 0x47 // Permanent error, deleted by originating SME.
	MmSmsDeliveryStateErrorDeletedByScAdministration MMSmsDeliveryState = 0x48 // Permanent error, deleted by SC administration.
	MmSmsDeliveryStateErrorMessageDoesNotExist       MMSmsDeliveryState = 0x49 // Permanent error, message does no longer exist.

	/* Temporary failures that became permanent */
	MmSmsDeliveryStateTemporaryFatalErrorCongestion        MMSmsDeliveryState = 0x60 // Permanent error, congestion.
	MmSmsDeliveryStateTemporaryFatalErrorSmeBusy           MMSmsDeliveryState = 0x61 // Permanent error, SME busy.
	MmSmsDeliveryStateTemporaryFatalErrorNoResponseFromSme MMSmsDeliveryState = 0x62 // Permanent error, no response from the SME.
	MmSmsDeliveryStateTemporaryFatalErrorServiceRejected   MMSmsDeliveryState = 0x63 // Permanent error, service rejected.
	MmSmsDeliveryStateTemporaryFatalErrorQosNotAvailable   MMSmsDeliveryState = 0x64 // Permanent error, QoS not available.
	MmSmsDeliveryStateTemporaryFatalErrorInSme             MMSmsDeliveryState = 0x65 // Permanent error in SME.

	/* Unknown, out of any possible valid value [0x00-0xFF] */
	MmSmsDeliveryStateUnknown MMSmsDeliveryState = 0x100 // Unknown state.

	/* --------------- 3GPP2 specific errors ---------------------- */

	/* Network problems */
	MmSmsDeliveryStateNetworkProblemAddressVacant             MMSmsDeliveryState = 0x200 // Permanent error in network, address vacant.
	MmSmsDeliveryStateNetworkProblemAddressTranslationFailure MMSmsDeliveryState = 0x201 // Permanent error in network, address translation failure.
	MmSmsDeliveryStateNetworkProblemNetworkResourceOutage     MMSmsDeliveryState = 0x202 // Permanent error in network, network resource outage.
	MmSmsDeliveryStateNetworkProblemNetworkFailure            MMSmsDeliveryState = 0x203 // Permanent error in network, network failure.
	MmSmsDeliveryStateNetworkProblemInvalidTeleserviceId      MMSmsDeliveryState = 0x204 // Permanent error in network, invalid teleservice id.
	MmSmsDeliveryStateNetworkProblemOther                     MMSmsDeliveryState = 0x205 // Permanent error, other network problem.
	/* Terminal problems */
	MmSmsDeliveryStateTerminalProblemNoPageResponse                   MMSmsDeliveryState = 0x220 // Permanent error in terminal, no page response.
	MmSmsDeliveryStateTerminalProblemDestinationBusy                  MMSmsDeliveryState = 0x221 // Permanent error in terminal, destination busy.
	MmSmsDeliveryStateTerminalProblemNoAcknowledgment                 MMSmsDeliveryState = 0x222 // Permanent error in terminal, no acknowledgement.
	MmSmsDeliveryStateTerminalProblemDestinationResourceShortage      MMSmsDeliveryState = 0x223 // Permanent error in terminal, destination resource shortage.
	MmSmsDeliveryStateTerminalProblemSmsDeliveryPostponed             MMSmsDeliveryState = 0x224 // Permanent error in terminal, SMS delivery postponed.
	MmSmsDeliveryStateTerminalProblemDestinationOutOfService          MMSmsDeliveryState = 0x225 // Permanent error in terminal, destination out of service.
	MmSmsDeliveryStateTerminalProblemDestinationNoLongerAtThisAddress MMSmsDeliveryState = 0x226 // Permanent error in terminal, destination no longer at this address.
	MmSmsDeliveryStateTerminalProblemOther                            MMSmsDeliveryState = 0x227 // Permanent error, other terminal problem.
	/* Radio problems */
	MmSmsDeliveryStateRadioInterfaceProblemResourceShortage MMSmsDeliveryState = 0x240 // Permanent error in radio interface, resource shortage.
	MmSmsDeliveryStateRadioInterfaceProblemIncompatibility  MMSmsDeliveryState = 0x241 // Permanent error in radio interface, problem incompatibility.
	MmSmsDeliveryStateRadioInterfaceProblemOther            MMSmsDeliveryState = 0x242 // Permanent error, other radio interface problem.
	/* General problems */
	MmSmsDeliveryStateGeneralProblemEncoding                         MMSmsDeliveryState = 0x260 // Permanent error, encoding.
	MmSmsDeliveryStateGeneralProblemSmsOriginationDenied             MMSmsDeliveryState = 0x261 // Permanent error, SMS origination denied.
	MmSmsDeliveryStateGeneralProblemSmsTerminationDenied             MMSmsDeliveryState = 0x262 // Permanent error, SMS termination denied.
	MmSmsDeliveryStateGeneralProblemSupplementaryServiceNotSupported MMSmsDeliveryState = 0x263 // Permanent error, supplementary service not supported.
	MmSmsDeliveryStateGeneralProblemSmsNotSupported                  MMSmsDeliveryState = 0x264 // Permanent error, SMS not supported.
	MmSmsDeliveryStateGeneralProblemMissingExpectedParameter         MMSmsDeliveryState = 0x266 // Permanent error, missing expected parameter.
	MmSmsDeliveryStateGeneralProblemMissingMandatoryParameter        MMSmsDeliveryState = 0x267 // Permanent error, missing mandatory parameter.
	MmSmsDeliveryStateGeneralProblemUnrecognizedParameterValue       MMSmsDeliveryState = 0x268 // Permanent error, unrecognized parameter value.
	MmSmsDeliveryStateGeneralProblemUnexpectedParameterValue         MMSmsDeliveryState = 0x269 // Permanent error, unexpected parameter value.
	MmSmsDeliveryStateGeneralProblemUserDataSizeError                MMSmsDeliveryState = 0x26A // Permanent error, user data size error.
	MmSmsDeliveryStateGeneralProblemOther                            MMSmsDeliveryState = 0x26B //  Permanent error, other general problem.

	/* Temporary network problems */
	MmSmsDeliveryStateTemporaryNetworkProblemAddressVacant             MMSmsDeliveryState = 0x300 // Temporary error in network, address vacant.
	MmSmsDeliveryStateTemporaryNetworkProblemAddressTranslationFailure MMSmsDeliveryState = 0x301 // Temporary error in network, address translation failure.
	MmSmsDeliveryStateTemporaryNetworkProblemNetworkResourceOutage     MMSmsDeliveryState = 0x302 // Temporary error in network, network resource outage.
	MmSmsDeliveryStateTemporaryNetworkProblemNetworkFailure            MMSmsDeliveryState = 0x303 // Temporary error in network, network failure.
	MmSmsDeliveryStateTemporaryNetworkProblemInvalidTeleserviceId      MMSmsDeliveryState = 0x304 // Temporary error in network, invalid teleservice id.
	MmSmsDeliveryStateTemporaryNetworkProblemOther                     MMSmsDeliveryState = 0x305 // Temporary error, other network problem.
	/* Temporary terminal problems */
	MmSmsDeliveryStateTemporaryTerminalProblemNoPageResponse                   MMSmsDeliveryState = 0x320 // Temporary error in terminal, no page response.
	MmSmsDeliveryStateTemporaryTerminalProblemDestinationBusy                  MMSmsDeliveryState = 0x321 // Temporary error in terminal, destination busy.
	MmSmsDeliveryStateTemporaryTerminalProblemNoAcknowledgment                 MMSmsDeliveryState = 0x322 // Temporary error in terminal, no acknowledgement.
	MmSmsDeliveryStateTemporaryTerminalProblemDestinationResourceShortage      MMSmsDeliveryState = 0x323 // Temporary error in terminal, destination resource shortage.
	MmSmsDeliveryStateTemporaryTerminalProblemSmsDeliveryPostponed             MMSmsDeliveryState = 0x324 // Temporary error in terminal, SMS delivery postponed.
	MmSmsDeliveryStateTemporaryTerminalProblemDestinationOutOfService          MMSmsDeliveryState = 0x325 // Temporary error in terminal, destination out of service.
	MmSmsDeliveryStateTemporaryTerminalProblemDestinationNoLongerAtThisAddress MMSmsDeliveryState = 0x326 // Temporary error in terminal, destination no longer at this address.
	MmSmsDeliveryStateTemporaryTerminalProblemOther                            MMSmsDeliveryState = 0x327 // Temporary error, other terminal problem.
	/* Temporary radio problems */
	MmSmsDeliveryStateTemporaryRadioInterfaceProblemResourceShortage MMSmsDeliveryState = 0x340 // Temporary error in radio interface, resource shortage.
	MmSmsDeliveryStateTemporaryRadioInterfaceProblemIncompatibility  MMSmsDeliveryState = 0x341 // Temporary error in radio interface, problem incompatibility.
	MmSmsDeliveryStateTemporaryRadioInterfaceProblemOther            MMSmsDeliveryState = 0x342 // Temporary error, other radio interface problem.
	/* Temporary general problems */
	MmSmsDeliveryStateTemporaryGeneralProblemEncoding                         MMSmsDeliveryState = 0x360 // Temporary error, encoding.
	MmSmsDeliveryStateTemporaryGeneralProblemSmsOriginationDenied             MMSmsDeliveryState = 0x361 // Temporary error, SMS origination denied.
	MmSmsDeliveryStateTemporaryGeneralProblemSmsTerminationDenied             MMSmsDeliveryState = 0x362 // Temporary error, SMS termination denied.
	MmSmsDeliveryStateTemporaryGeneralProblemSupplementaryServiceNotSupported MMSmsDeliveryState = 0x363 // Temporary error, supplementary service not supported.
	MmSmsDeliveryStateTemporaryGeneralProblemSmsNotSupported                  MMSmsDeliveryState = 0x364 // Temporary error, SMS not supported.
	MmSmsDeliveryStateTemporaryGeneralProblemMissingExpectedParameter         MMSmsDeliveryState = 0x366 // Temporary error, missing expected parameter.
	MmSmsDeliveryStateTemporaryGeneralProblemMissingMandatoryParameter        MMSmsDeliveryState = 0x367 // Temporary error, missing mandatory parameter.
	MmSmsDeliveryStateTemporaryGeneralProblemUnrecognizedParameterValue       MMSmsDeliveryState = 0x368 // Temporary error, unrecognized parameter value.
	MmSmsDeliveryStateTemporaryGeneralProblemUnexpectedParameterValue         MMSmsDeliveryState = 0x369 // Temporary error, unexpected parameter value.
	MmSmsDeliveryStateTemporaryGeneralProblemUserDataSizeError                MMSmsDeliveryState = 0x36A // Temporary error, user data size error.
	MmSmsDeliveryStateTemporaryGeneralProblemOther                            MMSmsDeliveryState = 0x36B // Temporary error, other general problem.

)

// MMSmsStorage Storage for SMS messages.
type MMSmsStorage uint32

//go:generate stringer -type=MMSmsStorage -trimprefix=MmSmsStorage
const (
	MmSmsStorageUnknown MMSmsStorage = 0 // Storage unknown.
	MmSmsStorageSm      MMSmsStorage = 1 // SIM card storage area.
	MmSmsStorageMe      MMSmsStorage = 2 // Mobile equipment storage area.
	MmSmsStorageMt      MMSmsStorage = 3 // Sum of SIM and Mobile equipment storages
	MmSmsStorageSr      MMSmsStorage = 4 // Status report message storage area.
	MmSmsStorageBm      MMSmsStorage = 5 // Broadcast message storage area.
	MmSmsStorageTa      MMSmsStorage = 6 // Terminal adaptor message storage area.

)

// MMSmsValidityType Type of SMS validity value.
type MMSmsValidityType uint32

//go:generate stringer -type=MMSmsValidityType -trimprefix=MmSmsValidityType
const (
	MmSmsValidityTypeUnknown  MMSmsValidityType = 0 // Validity type unknown.
	MmSmsValidityTypeRelative MMSmsValidityType = 1 // Relative validity.
	MmSmsValidityTypeAbsolute MMSmsValidityType = 2 // Absolute validity.
	MmSmsValidityTypeEnhanced MMSmsValidityType = 3 // Enhanced validity.

)

// MMSmsCdmaTeleserviceId Teleservice IDs supported for CDMA SMS, as defined in 3GPP2 X.S0004-550-E (section 2.256) and 3GPP2 C.S0015-B (section 3.4.3.1).
type MMSmsCdmaTeleserviceId uint32

//go:generate stringer -type=MMSmsCdmaTeleserviceId -trimprefix=MmSmsCdmaTeleserviceId
const (
	MmSmsCdmaTeleserviceIdUnknown MMSmsCdmaTeleserviceId = 0x0000 // Unknown.
	MmSmsCdmaTeleserviceIdCmt91   MMSmsCdmaTeleserviceId = 0x1000 // IS-91 Extended Protocol Enhanced Services.
	MmSmsCdmaTeleserviceIdWpt     MMSmsCdmaTeleserviceId = 0x1001 // Wireless Paging Teleservice.
	MmSmsCdmaTeleserviceIdWmt     MMSmsCdmaTeleserviceId = 0x1002 // Wireless Messaging Teleservice.
	MmSmsCdmaTeleserviceIdVmn     MMSmsCdmaTeleserviceId = 0x1003 // Voice Mail Notification.
	MmSmsCdmaTeleserviceIdWap     MMSmsCdmaTeleserviceId = 0x1004 // Wireless Application Protocol.
	MmSmsCdmaTeleserviceIdWemt    MMSmsCdmaTeleserviceId = 0x1005 // Wireless Enhanced Messaging Teleservice.
	MmSmsCdmaTeleserviceIdScpt    MMSmsCdmaTeleserviceId = 0x1006 // Service Category Programming Teleservice
	MmSmsCdmaTeleserviceIdCatpt   MMSmsCdmaTeleserviceId = 0x1007 // Card Application Toolkit Protocol Teleservice.

)

// MMSmsCdmaServiceCategory Service category for CDMA SMS, as defined in 3GPP2 C.R1001-D (section 9.3).
type MMSmsCdmaServiceCategory uint32

//go:generate stringer -type=MMSmsCdmaServiceCategory -trimprefix=MmSmsCdmaServiceCategory
const (
	MmSmsCdmaServiceCategoryUnknown                        MMSmsCdmaServiceCategory = 0x0000 // Unknown.
	MmSmsCdmaServiceCategoryEmergencyBroadcast             MMSmsCdmaServiceCategory = 0x0001 // Emergency broadcast.
	MmSmsCdmaServiceCategoryAdministrative                 MMSmsCdmaServiceCategory = 0x0002 // Administrative.
	MmSmsCdmaServiceCategoryMaintenance                    MMSmsCdmaServiceCategory = 0x0003 // Maintenance.
	MmSmsCdmaServiceCategoryGeneralNewsLocal               MMSmsCdmaServiceCategory = 0x0004 // General news (local).
	MmSmsCdmaServiceCategoryGeneralNewsRegional            MMSmsCdmaServiceCategory = 0x0005 // General news (regional).
	MmSmsCdmaServiceCategoryGeneralNewsNational            MMSmsCdmaServiceCategory = 0x0006 // General news (national).
	MmSmsCdmaServiceCategoryGeneralNewsInternational       MMSmsCdmaServiceCategory = 0x0007 // General news (international).
	MmSmsCdmaServiceCategoryBusinessNewsLocal              MMSmsCdmaServiceCategory = 0x0008 // Business/Financial news (local).
	MmSmsCdmaServiceCategoryBusinessNewsRegional           MMSmsCdmaServiceCategory = 0x0009 // Business/Financial news (regional).
	MmSmsCdmaServiceCategoryBusinessNewsNational           MMSmsCdmaServiceCategory = 0x000A // Business/Financial news (national).
	MmSmsCdmaServiceCategoryBusinessNewsInternational      MMSmsCdmaServiceCategory = 0x000B // Business/Financial news (international).
	MmSmsCdmaServiceCategorySportsNewsLocal                MMSmsCdmaServiceCategory = 0x000C // Sports news (local).
	MmSmsCdmaServiceCategorySportsNewsRegional             MMSmsCdmaServiceCategory = 0x000D // Sports news (regional).
	MmSmsCdmaServiceCategorySportsNewsNational             MMSmsCdmaServiceCategory = 0x000E // Sports news (national).
	MmSmsCdmaServiceCategorySportsNewsInternational        MMSmsCdmaServiceCategory = 0x000F // Sports news (international).
	MmSmsCdmaServiceCategoryEntertainmentNewsLocal         MMSmsCdmaServiceCategory = 0x0010 // Entertainment news (local).
	MmSmsCdmaServiceCategoryEntertainmentNewsRegional      MMSmsCdmaServiceCategory = 0x0011 // Entertainment news (regional).
	MmSmsCdmaServiceCategoryEntertainmentNewsNational      MMSmsCdmaServiceCategory = 0x0012 // Entertainment news (national).
	MmSmsCdmaServiceCategoryEntertainmentNewsInternational MMSmsCdmaServiceCategory = 0x0013 // Entertainment news (international).
	MmSmsCdmaServiceCategoryLocalWeather                   MMSmsCdmaServiceCategory = 0x0014 // Local weather.
	MmSmsCdmaServiceCategoryTrafficReport                  MMSmsCdmaServiceCategory = 0x0015 // Area traffic report.
	MmSmsCdmaServiceCategoryFlightSchedules                MMSmsCdmaServiceCategory = 0x0016 // Local airport flight schedules.
	MmSmsCdmaServiceCategoryRestaurants                    MMSmsCdmaServiceCategory = 0x0017 // Restaurants.
	MmSmsCdmaServiceCategoryLodgings                       MMSmsCdmaServiceCategory = 0x0018 // Lodgings.
	MmSmsCdmaServiceCategoryRetailDirectory                MMSmsCdmaServiceCategory = 0x0019 // Retail directory.
	MmSmsCdmaServiceCategoryAdvertisements                 MMSmsCdmaServiceCategory = 0x001A // Advertisements.
	MmSmsCdmaServiceCategoryStockQuotes                    MMSmsCdmaServiceCategory = 0x001B // Stock quotes.
	MmSmsCdmaServiceCategoryEmployment                     MMSmsCdmaServiceCategory = 0x001C // Employment.
	MmSmsCdmaServiceCategoryHospitals                      MMSmsCdmaServiceCategory = 0x001D // Medical / Health / Hospitals.
	MmSmsCdmaServiceCategoryTechnologyNews                 MMSmsCdmaServiceCategory = 0x001E // Technology news.
	MmSmsCdmaServiceCategoryMulticategory                  MMSmsCdmaServiceCategory = 0x001F // Multi-category.
	MmSmsCdmaServiceCategoryCmasPresidentialAlert          MMSmsCdmaServiceCategory = 0x1000 // Presidential alert.
	MmSmsCdmaServiceCategoryCmasExtremeThreat              MMSmsCdmaServiceCategory = 0x1001 // Extreme threat.
	MmSmsCdmaServiceCategoryCmasSevereThreat               MMSmsCdmaServiceCategory = 0x1002 // Severe threat.
	MmSmsCdmaServiceCategoryCmasChildAbductionEmergency    MMSmsCdmaServiceCategory = 0x1003 // Child abduction emergency.
	MmSmsCdmaServiceCategoryCmasTest                       MMSmsCdmaServiceCategory = 0x1004 // CMAS test.

)

// MMModemLocationSource Sources of location information supported by the modem.
type MMModemLocationSource uint32

//go:generate stringer -type=MMModemLocationSource -trimprefix=MmModemLocationSource
const (
	MmModemLocationSourceNone         MMModemLocationSource = 0      // None.
	MmModemLocationSource3gppLacCi    MMModemLocationSource = 1 << 0 //  Location Area Code and Cell ID.
	MmModemLocationSourceGpsRaw       MMModemLocationSource = 1 << 1 // GPS location given by predefined keys.
	MmModemLocationSourceGpsNmea      MMModemLocationSource = 1 << 2 // GPS location given as NMEA traces.
	MmModemLocationSourceCdmaBs       MMModemLocationSource = 1 << 3 // CDMA base station position.
	MmModemLocationSourceGpsUnmanaged MMModemLocationSource = 1 << 4 // No location given, just GPS module setup.
	MmModemLocationSourceAgpsMsa      MMModemLocationSource = 1 << 5 // Mobile Station Assisted A-GPS location requested.
	MmModemLocationSourceAgpsMsb      MMModemLocationSource = 1 << 6 // Mobile Station Based A-GPS location requested.

)

// GetAllSources returns all sources
func (ls MMModemLocationSource) GetAllSources() []MMModemLocationSource {

	return []MMModemLocationSource{MmModemLocationSource3gppLacCi, MmModemLocationSourceGpsRaw,
		MmModemLocationSourceGpsNmea, MmModemLocationSourceCdmaBs, MmModemLocationSourceGpsUnmanaged,
		MmModemLocationSourceAgpsMsa, MmModemLocationSourceAgpsMsb}
}

// BitmaskToSlice bitmask to slice
func (ls MMModemLocationSource) BitmaskToSlice(bitmask uint32) (sources []MMModemLocationSource) {
	if bitmask == 0 {
		return
	}
	for idx, x := range ls.GetAllSources() {
		if bitmask&(1<<idx) > 0 {
			sources = append(sources, x)
		}
	}
	return sources
}

// SliceToBitmask slice to bitmask
func (ls MMModemLocationSource) SliceToBitmask(sources []MMModemLocationSource) (bitmask uint32) {
	bitmask = 0
	for idx, x := range ls.GetAllSources() {
		for _, y := range sources {
			if x == y {
				bitmask = bitmask | (1 << idx)
			}
		}
	}
	return bitmask
}

// MMModemLocationAssistanceDataType Type of assistance data that may be injected to the GNSS module.
type MMModemLocationAssistanceDataType uint32

//go:generate stringer -type=MMModemLocationAssistanceDataType -trimprefix=MmModemLocationAssistanceDataType
const (
	MmModemLocationAssistanceDataTypeNone MMModemLocationAssistanceDataType = 0      // None.
	MmModemLocationAssistanceDataTypeXtra MMModemLocationAssistanceDataType = 1 << 0 // Qualcomm gpsOneXTRA.
)

// GetAllAssistanceData returns all assistance data
func (ad MMModemLocationAssistanceDataType) GetAllAssistanceData() []MMModemLocationAssistanceDataType {

	return []MMModemLocationAssistanceDataType{MmModemLocationAssistanceDataTypeXtra}
}

// BitmaskToSlice bitmask to slice
func (ad MMModemLocationAssistanceDataType) BitmaskToSlice(bitmask uint32) (data []MMModemLocationAssistanceDataType) {
	if bitmask == 0 {
		return
	}
	for idx, x := range ad.GetAllAssistanceData() {
		if bitmask&(1<<idx) > 0 {
			data = append(data, x)
		}
	}
	return
}

// SliceToBitmask slice to bitmask
func (ad MMModemLocationAssistanceDataType) SliceToBitmask(data []MMModemLocationAssistanceDataType) (bitmask uint32) {
	bitmask = 0
	for idx, x := range ad.GetAllAssistanceData() {
		for _, y := range data {
			if x == y {
				bitmask = bitmask | (1 << idx)
			}
		}
	}
	return
}

// MMModemContactsStorage Specifies different storage locations for contact information.
type MMModemContactsStorage uint32

//go:generate stringer -type=MMModemContactsStorage -trimprefix=MmModemContactsStorage
const (
	MmModemContactsStorageUnknown MMModemContactsStorage = 0 // Unknown location.
	MmModemContactsStorageMe      MMModemContactsStorage = 1 // Device's local memory.
	MmModemContactsStorageSm      MMModemContactsStorage = 2 // Card inserted in the device (like a SIM/RUIM).
	MmModemContactsStorageMt      MMModemContactsStorage = 3 // Combined device/ME and SIM/SM phonebook.

)

// MMBearerType Type of context (2G/3G) or bearer (4G).
type MMBearerType uint32

//go:generate stringer -type=MMBearerType -trimprefix=MmBearerType
const (
	MmBearerTypeUnknown       MMBearerType = 0 // Unknown bearer.
	MmBearerTypeDefault       MMBearerType = 1 // Primary context (2G/3G) or default bearer (4G), defined by the user of the API.
	MmBearerTypeDefaultAttach MMBearerType = 2 // The initial default bearer established during LTE attach procedure, automatically connected as long as the device is  registered in the LTE network.
	MmBearerTypeDedicated     MMBearerType = 3 // Secondary context (2G/3G) or dedicated bearer (4G), defined by the user of the API. These bearers use the same IP address  used by a primary context or default bearer and provide a dedicated flow for  specific traffic with different QoS settings.
)

// MMBearerIpMethod Type of IP method configuration to be used in a given Bearer.
type MMBearerIpMethod uint32

//go:generate stringer -type=MMBearerIpMethod -trimprefix=MmBearerIpMethod
const (
	MmBearerIpMethodUnknown MMBearerIpMethod = 0 // Unknown method.
	MmBearerIpMethodPpp     MMBearerIpMethod = 1 //  Use PPP to get IP addresses and DNS information. For IPv6, use PPP to retrieve the 64-bit Interface Identifier, use the IID to construct an IPv6 link-local address by following RFC 5072, and then run DHCP over the PPP link to retrieve DNS settings.
	MmBearerIpMethodStatic  MMBearerIpMethod = 2 // Use the provided static IP configuration given by the modem to configure the IP data interface.  Note that DNS servers may not be provided by the network or modem firmware.
	MmBearerIpMethodDhcp    MMBearerIpMethod = 3 // Begin DHCP or IPv6 SLAAC on the data interface to obtain any necessary IP configuration details that are not already provided by the IP configuration.  For IPv4 bearers DHCP should be used.  For IPv6 bearers SLAAC should be used, and the IP configuration may already contain a link-local address that should be assigned to the interface before SLAAC is started to obtain the rest of the configuration.

)

// MMBearerIpFamily Type of IP family to be used in a given Bearer.
type MMBearerIpFamily uint32

//go:generate stringer -type=MMBearerIpFamily -trimprefix=MmBearerIpFamily
const (
	MmBearerIpFamilyNone   MMBearerIpFamily = 0          // None or unknown.
	MmBearerIpFamilyIpv4   MMBearerIpFamily = 1 << 0     // IPv4.
	MmBearerIpFamilyIpv6   MMBearerIpFamily = 1 << 1     // IPv6.
	MmBearerIpFamilyIpv4v6 MMBearerIpFamily = 1 << 2     // IPv4 and IPv6.
	MmBearerIpFamilyAny    MMBearerIpFamily = 0xFFFFFFFF // Mask specifying all IP families.

)

// GetAllIPFamilies returns all ip families
func (i MMBearerIpFamily) GetAllIPFamilies() []MMBearerIpFamily {

	return []MMBearerIpFamily{MmBearerIpFamilyIpv4, MmBearerIpFamilyIpv6,
		MmBearerIpFamilyIpv4v6}
}

// BitmaskToSlice bitmask to slice
func (i MMBearerIpFamily) BitmaskToSlice(bitmask uint32) (ipFamilies []MMBearerIpFamily) {
	if bitmask == 0 {
		return
	}
	for idx, x := range i.GetAllIPFamilies() {
		if bitmask&(1<<idx) > 0 {
			ipFamilies = append(ipFamilies, x)
		}
	}
	return ipFamilies
}

// SliceToBitmask slice to bitmask
func (i MMBearerIpFamily) SliceToBitmask(ipFamilies []MMBearerIpFamily) (bitmask uint32) {
	bitmask = 0
	for idx, x := range i.GetAllIPFamilies() {
		for _, y := range ipFamilies {
			if x == y {
				bitmask = bitmask | (1 << idx)
			}
		}
	}
	return bitmask
}

// MMBearerAllowedAuth Allowed authentication methods when authenticating with the network.
type MMBearerAllowedAuth uint32

//go:generate stringer -type=MMBearerAllowedAuth -trimprefix=MmBearerAllowedAuth
const (
	MmBearerAllowedAuthUnknown MMBearerAllowedAuth = 0 // Unknown.
	/* bits 0..4 order match Ericsson device bitmap */
	MmBearerAllowedAuthNone     MMBearerAllowedAuth = 1 << 0 // None.
	MmBearerAllowedAuthPap      MMBearerAllowedAuth = 1 << 1 // PAP.
	MmBearerAllowedAuthChap     MMBearerAllowedAuth = 1 << 2 // CHAP.
	MmBearerAllowedAuthMschap   MMBearerAllowedAuth = 1 << 3 // MS-CHAP.
	MmBearerAllowedAuthMschapv2 MMBearerAllowedAuth = 1 << 4 // MS-CHAP v2.
	MmBearerAllowedAuthEap      MMBearerAllowedAuth = 1 << 5 // EAP.

)

// MMModemCdmaRegistrationState Registration state of a CDMA modem.
type MMModemCdmaRegistrationState uint32

//go:generate stringer -type=MMModemCdmaRegistrationState -trimprefix=MmModemCdmaRegistrationState
const (
	MmModemCdmaRegistrationStateUnknown    MMModemCdmaRegistrationState = 0 // Registration status is unknown or the device is not registered.
	MmModemCdmaRegistrationStateRegistered MMModemCdmaRegistrationState = 1 // Registered, but roaming status is unknown or cannot be provided by the device. The device may or may not be roaming.
	MmModemCdmaRegistrationStateHome       MMModemCdmaRegistrationState = 2 // Currently registered on the home network.
	MmModemCdmaRegistrationStateRoaming    MMModemCdmaRegistrationState = 3 // Currently registered on a roaming network.

)

// MMModemCdmaActivationState Activation state of a CDMA modem.
type MMModemCdmaActivationState uint32

//go:generate stringer -type=MMModemCdmaActivationState -trimprefix=MmModemCdmaActivationState
const (
	MmModemCdmaActivationStateUnknown            MMModemCdmaActivationState = 0 // Unknown activation state.
	MmModemCdmaActivationStateNotActivated       MMModemCdmaActivationState = 1 // Device is not activated
	MmModemCdmaActivationStateActivating         MMModemCdmaActivationState = 2 // Device is activating
	MmModemCdmaActivationStatePartiallyActivated MMModemCdmaActivationState = 3 // Device is partially activated; carrier-specific steps required to continue.
	MmModemCdmaActivationStateActivated          MMModemCdmaActivationState = 4 // Device is ready for use.

)

// MMModemCdmaRmProtocol Protocol of the Rm interface in modems with CDMA capabilities.
type MMModemCdmaRmProtocol uint32

//go:generate stringer -type=MMModemCdmaRmProtocol -trimprefix=MmModemCdmaRmProtocol
const (
	MmModemCdmaRmProtocolUnknown           MMModemCdmaRmProtocol = 0 // Unknown protocol.
	MmModemCdmaRmProtocolAsync             MMModemCdmaRmProtocol = 1 // Asynchronous data or fax.
	MmModemCdmaRmProtocolPacketRelay       MMModemCdmaRmProtocol = 2 // Packet data service, Relay Layer Rm interface.
	MmModemCdmaRmProtocolPacketNetworkPpp  MMModemCdmaRmProtocol = 3 // Packet data service, Network Layer Rm interface, PPP.
	MmModemCdmaRmProtocolPacketNetworkSlip MMModemCdmaRmProtocol = 4 // Packet data service, Network Layer Rm interface, SLIP.
	MmModemCdmaRmProtocolStuIii            MMModemCdmaRmProtocol = 5 // STU-III service.

)

// MMModem3gppRegistrationState GSM registration code as defined in 3GPP TS 27.007.
type MMModem3gppRegistrationState uint32

//go:generate stringer -type=MMModem3gppRegistrationState -trimprefix=MmModem3gppRegistrationState
const (
	MmModem3gppRegistrationStateIdle                    MMModem3gppRegistrationState = 0  // Not registered, not searching for new operator to register.
	MmModem3gppRegistrationStateHome                    MMModem3gppRegistrationState = 1  // Registered on home network.
	MmModem3gppRegistrationStateSearching               MMModem3gppRegistrationState = 2  // Not registered, searching for new operator to register with.
	MmModem3gppRegistrationStateDenied                  MMModem3gppRegistrationState = 3  // Registration denied.
	MmModem3gppRegistrationStateUnknown                 MMModem3gppRegistrationState = 4  // Unknown registration status.
	MmModem3gppRegistrationStateRoaming                 MMModem3gppRegistrationState = 5  // Registered on a roaming network.
	MmModem3gppRegistrationStateHomeSmsOnly             MMModem3gppRegistrationState = 6  // Registered for "SMS only", home network (applicable only when on LTE).
	MmModem3gppRegistrationStateRoamingSmsOnly          MMModem3gppRegistrationState = 7  // Registered for "SMS only", roaming network (applicable only when on LTE).
	MmModem3gppRegistrationStateEmergencyOnly           MMModem3gppRegistrationState = 8  // Emergency services only.
	MmModem3gppRegistrationStateHomeCsfbNotPreferred    MMModem3gppRegistrationState = 9  // Registered for "CSFB not preferred", home network (applicable only when on LTE).
	MmModem3gppRegistrationStateRoamingCsfbNotPreferred MMModem3gppRegistrationState = 10 // Registered for "CSFB not preferred", roaming network (applicable only when on LTE).

)

// MMModem3gppFacility A bitfield describing which facilities have a lock enabled, i.e., requires a pin or unlock code. The facilities include the personalizations (device locks) described in 3GPP spec TS 22.022, and the PIN and PIN2 locks, which are SIM locks.
type MMModem3gppFacility uint32

//go:generate stringer -type=MMModem3gppFacility -trimprefix=MmModem3gppFacility
const (
	MmModem3gppFacilityNone         MMModem3gppFacility = 0      // No facility.
	MmModem3gppFacilitySim          MMModem3gppFacility = 1 << 0 // SIM lock.
	MmModem3gppFacilityFixedDialing MMModem3gppFacility = 1 << 1 // Fixed dialing (PIN2) SIM lock.
	MmModem3gppFacilityPhSim        MMModem3gppFacility = 1 << 2 // Device is locked to a specific SIM.
	MmModem3gppFacilityPhFsim       MMModem3gppFacility = 1 << 3 // Device is locked to first SIM inserted.
	MmModem3gppFacilityNetPers      MMModem3gppFacility = 1 << 4 // Network personalization.
	MmModem3gppFacilityNetSubPers   MMModem3gppFacility = 1 << 5 // Network subset personalization.
	MmModem3gppFacilityProviderPers MMModem3gppFacility = 1 << 6 // Service provider personalization.
	MmModem3gppFacilityCorpPers     MMModem3gppFacility = 1 << 7 // Corporate personalization.

)

// GetAllFacilities returns all facilities
func (f MMModem3gppFacility) GetAllFacilities() []MMModem3gppFacility {

	return []MMModem3gppFacility{MmModem3gppFacilitySim,
		MmModem3gppFacilityFixedDialing, MmModem3gppFacilityPhSim, MmModem3gppFacilityNetPers,
		MmModem3gppFacilityNetSubPers, MmModem3gppFacilityProviderPers, MmModem3gppFacilityCorpPers}
}

// BitmaskToSlice bitmask to slice
func (f MMModem3gppFacility) BitmaskToSlice(bitmask uint32) (facilities []MMModem3gppFacility) {
	if bitmask == 0 {
		return
	}
	for idx, x := range f.GetAllFacilities() {
		if bitmask&(1<<idx) > 0 {
			facilities = append(facilities, x)
		}
	}
	return facilities
}

// SliceToBitmask slice to bitmask
func (f MMModem3gppFacility) SliceToBitmask(facilities []MMModem3gppFacility) (bitmask uint32) {
	bitmask = 0
	for idx, x := range f.GetAllFacilities() {
		for _, y := range facilities {
			if x == y {
				bitmask = bitmask | (1 << idx)
			}
		}
	}
	return bitmask
}

// MMModem3gppNetworkAvailability Network availability status as defined in 3GPP TS 27.007 section 7.3.
type MMModem3gppNetworkAvailability uint32

//go:generate stringer -type=MMModem3gppNetworkAvailability -trimprefix=MmModem3gppNetworkAvailability
const (
	MmModem3gppNetworkAvailabilityUnknown   MMModem3gppNetworkAvailability = 0 // Unknown availability.
	MmModem3gppNetworkAvailabilityAvailable MMModem3gppNetworkAvailability = 1 // Network is available.
	MmModem3gppNetworkAvailabilityCurrent   MMModem3gppNetworkAvailability = 2 // Network is the current one.
	MmModem3gppNetworkAvailabilityForbidden MMModem3gppNetworkAvailability = 3 // Network is forbidden.

)

// MMModem3gppSubscriptionState Describes the current subscription status of the SIM.  This value is only available after the modem attempts to register with the network.
type MMModem3gppSubscriptionState uint32

//go:generate stringer -type=MMModem3gppSubscriptionState -trimprefix=MmModem3gppSubscriptionState
const (
	MmModem3gppSubscriptionStateUnknown       MMModem3gppSubscriptionState = 0 // The subscription state is unknown.
	MmModem3gppSubscriptionStateUnprovisioned MMModem3gppSubscriptionState = 1 // The account is unprovisioned.
	MmModem3gppSubscriptionStateProvisioned   MMModem3gppSubscriptionState = 2 // The account is provisioned and has data available.
	MmModem3gppSubscriptionStateOutOfData     MMModem3gppSubscriptionState = 3 // The account is provisioned but there is no data left.

)

// MMModem3gppUssdSessionState State of a USSD session.
type MMModem3gppUssdSessionState uint32

//go:generate stringer -type=MMModem3gppUssdSessionState -trimprefix=MmModem3gppUssdSessionState
const (
	MmModem3gppUssdSessionStateUnknown      MMModem3gppUssdSessionState = 0 // Unknown state.
	MmModem3gppUssdSessionStateIdle         MMModem3gppUssdSessionState = 1 // No active session.
	MmModem3gppUssdSessionStateActive       MMModem3gppUssdSessionState = 2 // A session is active and the mobile is waiting for a response.
	MmModem3gppUssdSessionStateUserResponse MMModem3gppUssdSessionState = 3 // The network is waiting for the client's response.

)

// MMModem3gppEpsUeModeOperation UE mode of operation for EPS, as per 3GPP TS 24.301.
type MMModem3gppEpsUeModeOperation uint32

//go:generate stringer -type=MMModem3gppEpsUeModeOperation -trimprefix=MmModem3gppEpsUeModeOperation
const (
	MmModem3gppEpsUeModeOperationUnknown MMModem3gppEpsUeModeOperation = 0 //  Unknown or not applicable.
	MmModem3gppEpsUeModeOperationPs1     MMModem3gppEpsUeModeOperation = 1 // PS mode 1 of operation: EPS only, voice-centric.
	MmModem3gppEpsUeModeOperationPs2     MMModem3gppEpsUeModeOperation = 2 // PS mode 2 of operation: EPS only, data-centric.
	MmModem3gppEpsUeModeOperationCsps1   MMModem3gppEpsUeModeOperation = 3 // CS/PS mode 1 of operation: EPS and non-EPS, voice-centric.
	MmModem3gppEpsUeModeOperationCsps2   MMModem3gppEpsUeModeOperation = 4 // CS/PS mode 2 of operation: EPS and non-EPS, data-centric.

)

// MMFirmwareImageType Type of firmware image.
type MMFirmwareImageType uint32

//go:generate stringer -type=MMFirmwareImageType  -trimprefix=MmFirmwareImageType
const (
	MmFirmwareImageTypeUnknown MMFirmwareImageType = 0 // Unknown firmware type.
	MmFirmwareImageTypeGeneric MMFirmwareImageType = 1 // Generic firmware image.
	MmFirmwareImageTypeGobi    MMFirmwareImageType = 2 // Firmware image of Gobi devices.

)

// MMOmaFeature Features that can be enabled or disabled in the OMA device management support.
type MMOmaFeature uint32

//go:generate stringer -type=MMOmaFeature -trimprefix=MmOmaFeature
const (
	MmOmaFeatureNone                MMOmaFeature = 0      // None.
	MmOmaFeatureDeviceProvisioning  MMOmaFeature = 1 << 0 // Device provisioning service.
	MmOmaFeaturePrlUpdate           MMOmaFeature = 1 << 1 // PRL update service.
	MmOmaFeatureHandsFreeActivation MMOmaFeature = 1 << 2 // Hands free activation service.

)

// GetAllFeatures returns all features
func (mmo MMOmaFeature) GetAllFeatures() []MMOmaFeature {

	return []MMOmaFeature{MmOmaFeatureDeviceProvisioning, MmOmaFeaturePrlUpdate,
		MmOmaFeatureHandsFreeActivation}
}

// BitmaskToSlice bitmask to slice
func (mmo MMOmaFeature) BitmaskToSlice(bitmask uint32) (features []MMOmaFeature) {
	if bitmask == 0 {
		return
	}
	for idx, x := range mmo.GetAllFeatures() {
		if bitmask&(1<<idx) > 0 {
			features = append(features, x)
		}
	}
	return features
}

// SliceToBitmask slice to bitmask
func (mmo MMOmaFeature) SliceToBitmask(features []MMOmaFeature) (bitmask uint32) {
	bitmask = 0
	for idx, x := range mmo.GetAllFeatures() {
		for _, y := range features {
			if x == y {
				bitmask = bitmask | (1 << idx)
			}
		}
	}
	return bitmask
}

// MMOmaSessionType Type of OMA device management session.
type MMOmaSessionType uint32

//go:generate stringer -type=MMOmaSessionType -trimprefix=MmOmaSessionType
const (
	MmOmaSessionTypeUnknown                            MMOmaSessionType = 0  // Unknown session type.
	MmOmaSessionTypeClientInitiatedDeviceConfigure     MMOmaSessionType = 10 // Client-initiated device configure.
	MmOmaSessionTypeClientInitiatedPrlUpdate           MMOmaSessionType = 11 // Client-initiated PRL update.
	MmOmaSessionTypeClientInitiatedHandsFreeActivation MMOmaSessionType = 12 // Client-initiated hands free activation.
	MmOmaSessionTypeNetworkInitiatedDeviceConfigure    MMOmaSessionType = 20 // Network-initiated device configure.
	MmOmaSessionTypeNetworkInitiatedPrlUpdate          MMOmaSessionType = 21 // Network-initiated PRL update.
	MmOmaSessionTypeDeviceInitiatedPrlUpdate           MMOmaSessionType = 30 // Device-initiated PRL update.
	MmOmaSessionTypeDeviceInitiatedHandsFreeActivation MMOmaSessionType = 31 // Device-initiated hands free activation.

)

// MMOmaSessionState State of the OMA device management session.
type MMOmaSessionState int32

//go:generate stringer -type=MMOmaSessionState -trimprefix=MmOmaSessionState
const (
	MmOmaSessionStateFailed               MMOmaSessionState = -1 // Failed.
	MmOmaSessionStateUnknown              MMOmaSessionState = 0  // Unknown.
	MmOmaSessionStateStarted              MMOmaSessionState = 1  // Started.
	MmOmaSessionStateRetrying             MMOmaSessionState = 2  // Retrying.
	MmOmaSessionStateConnecting           MMOmaSessionState = 3  // Connecting.
	MmOmaSessionStateConnected            MMOmaSessionState = 4  // Connected.
	MmOmaSessionStateAuthenticated        MMOmaSessionState = 5  // Authenticated.
	MmOmaSessionStateMdnDownloaded        MMOmaSessionState = 10 // MDN downloaded.
	MmOmaSessionStateMsidDownloaded       MMOmaSessionState = 11 // MSID downloaded.
	MmOmaSessionStatePrlDownloaded        MMOmaSessionState = 12 // PRL downloaded.
	MmOmaSessionStateMipProfileDownloaded MMOmaSessionState = 13 // MIP profile downloaded.
	MmOmaSessionStateCompleted            MMOmaSessionState = 20 // Session completed.

)

// MMOmaSessionStateFailedReason Reason of failure in the OMA device management session.
type MMOmaSessionStateFailedReason uint32

//go:generate stringer -type=MMOmaSessionStateFailedReason -trimprefix=MmOmaSessionStateFailedReason
const (
	MmOmaSessionStateFailedReasonUnknown              MMOmaSessionStateFailedReason = 0 // No reason or unknown.
	MmOmaSessionStateFailedReasonNetworkUnavailable   MMOmaSessionStateFailedReason = 1 // Network unavailable.
	MmOmaSessionStateFailedReasonServerUnavailable    MMOmaSessionStateFailedReason = 2 // Server unavailable.
	MmOmaSessionStateFailedReasonAuthenticationFailed MMOmaSessionStateFailedReason = 3 // Authentication failed.
	MmOmaSessionStateFailedReasonMaxRetryExceeded     MMOmaSessionStateFailedReason = 4 // Maximum retries exceeded.
	MmOmaSessionStateFailedReasonSessionCancelled     MMOmaSessionStateFailedReason = 5 // Session cancelled.

)

// MMCallState State of Call.
type MMCallState int32

//go:generate stringer -type=MMCallState  -trimprefix=MmCallState
const (
	MmCallStateUnknown    MMCallState = 0 // default state for a new outgoing call.
	MmCallStateDialing    MMCallState = 1 // outgoing call started. Wait for free channel.
	MmCallStateRingingOut MMCallState = 2 // incoming call is waiting for an answer.
	MmCallStateRingingIn  MMCallState = 3 // outgoing call attached to GSM network, waiting for an answer.
	MmCallStateActive     MMCallState = 4 // call is active between two peers.
	MmCallStateHeld       MMCallState = 5 // held call (by +CHLD AT command).
	MmCallStateWaiting    MMCallState = 6 // waiting call (by +CCWA AT command).
	MmCallStateTerminated MMCallState = 7 // call is terminated.

)

// MMCallStateReason Reason for the state change in the call.
type MMCallStateReason int32

//go:generate stringer -type=MMCallStateReason -trimprefix=MmCallStateReason
const (
	MmCallStateReasonUnknown          MMCallStateReason = 0 // Default value for a new outgoing call.
	MmCallStateReasonOutgoingStarted  MMCallStateReason = 1 // Outgoing call is started.
	MmCallStateReasonIncomingNew      MMCallStateReason = 2 // Received a new incoming call.
	MmCallStateReasonAccepted         MMCallStateReason = 3 // Dialing or Ringing call is accepted.
	MmCallStateReasonTerminated       MMCallStateReason = 4 // Call is correctly terminated.
	MmCallStateReasonRefusedOrBusy    MMCallStateReason = 5 // Remote peer is busy or refused call.
	MmCallStateReasonError            MMCallStateReason = 6 // Wrong number or generic network error.
	MmCallStateReasonAudioSetupFailed MMCallStateReason = 7 // Error setting up audio channel.
	MmCallStateReasonTransferred      MMCallStateReason = 8 // Call has been transferred.
	MmCallStateReasonDeflected        MMCallStateReason = 9 // Call has been deflected to a new number.
)

// MMCallDirection Direction of the call.
type MMCallDirection int32

//go:generate stringer -type=MMCallDirection -trimprefix=MmCallDirection
const (
	MmCallDirectionUnknown  MMCallDirection = 0 // unknown.
	MmCallDirectionIncoming MMCallDirection = 1 // call from network.
	MmCallDirectionOutgoing MMCallDirection = 2 // call to network.

)

// MMModemFirmwareUpdateMethod Type of firmware update method supported by the module.
type MMModemFirmwareUpdateMethod uint32

//go:generate stringer -type=MMModemFirmwareUpdateMethod -trimprefix=MmModemFirmwareUpdateMethod
const (
	MmModemFirmwareUpdateMethodNone     MMModemFirmwareUpdateMethod = 0      // No method specified.
	MmModemFirmwareUpdateMethodFastboot MMModemFirmwareUpdateMethod = 1 << 0 // Device supports fastboot-based update.
	MmModemFirmwareUpdateMethodQmiPdc   MMModemFirmwareUpdateMethod = 1 << 1 // Device supports QMI PDC based update.

)

// GetAllUpdateMethods returns all update methods
func (fu MMModemFirmwareUpdateMethod) GetAllUpdateMethods() []MMModemFirmwareUpdateMethod {

	return []MMModemFirmwareUpdateMethod{MmModemFirmwareUpdateMethodFastboot, MmModemFirmwareUpdateMethodQmiPdc}
}

// BitmaskToSlice bitmask to slice
func (fu MMModemFirmwareUpdateMethod) BitmaskToSlice(bitmask uint32) (ipFamilies []MMModemFirmwareUpdateMethod) {
	if bitmask == 0 {
		return
	}
	for idx, x := range fu.GetAllUpdateMethods() {
		if bitmask&(1<<idx) > 0 {
			ipFamilies = append(ipFamilies, x)
		}
	}
	return ipFamilies
}

// SliceToBitmask slice to bitmask
func (fu MMModemFirmwareUpdateMethod) SliceToBitmask(updateMethods []MMModemFirmwareUpdateMethod) (bitmask uint32) {
	bitmask = 0
	for idx, x := range fu.GetAllUpdateMethods() {
		for _, y := range updateMethods {
			if x == y {
				bitmask = bitmask | (1 << idx)
			}
		}
	}
	return bitmask
}

// MMLoggingLevel Logging Level of ModemManager
type MMLoggingLevel string

// multiple logging levels for modem manager
const (
	MMLoggingLevelError   MMLoggingLevel = "ERR"   // logging level error.
	MMLoggingLevelWarning MMLoggingLevel = "WARN"  // logging level warning.
	MMLoggingLevelDebug   MMLoggingLevel = "DEBUG" // logging level debug.

)

// MMKernelPropertyAction The type of action, given as a string value (signature "s"). This parameter is MANDATORY.
type MMKernelPropertyAction string

// mm kernel property actions
const (
	MMKernelPropertyActionAdd    MMKernelPropertyAction = "add"    // A new kernel device has been added.
	MMKernelPropertyActionRemove MMKernelPropertyAction = "remove" // An existing kernel device has been removed.
)

/* Errors */

// MMCoreError Common errors that may be reported by ModemManager.
type MMCoreError uint32

//go:generate stringer -type=MMCoreError -trimprefix=MMCoreError
const (
	MmCoreErrorFailed       MMCoreError = 0  // Operation failed.
	MmCoreErrorCancelled    MMCoreError = 1  // Operation was cancelled.
	MmCoreErrorAborted      MMCoreError = 2  // Operation was aborted.
	MmCoreErrorUnsupported  MMCoreError = 3  // Operation is not supported.
	MmCoreErrorNoPlugins    MMCoreError = 4  // Cannot operate without valid plugins.
	MmCoreErrorUnauthorized MMCoreError = 5  // Authorization is required to perform the operation.
	MmCoreErrorInvalidArgs  MMCoreError = 6  // Invalid arguments given.
	MmCoreErrorInProgress   MMCoreError = 7  // Operation is already in progress.
	MmCoreErrorWrongState   MMCoreError = 8  // Operation cannot be executed in the current state.
	MmCoreErrorConnected    MMCoreError = 9  // Operation cannot be executed while being connected.
	MmCoreErrorTooMany      MMCoreError = 10 // Too many items.
	MmCoreErrorNotFound     MMCoreError = 11 // Item not found.
	MmCoreErrorRetry        MMCoreError = 12 // Operation cannot yet be performed, retry later.
	MmCoreErrorExists       MMCoreError = 13 // Item already exists.
)

// MMMobileEquipmentError Enumeration of Mobile Equipment errors, as defined in 3GPP TS 07.07 version 7.8.0.
type MMMobileEquipmentError uint32

//go:generate stringer -type=MMMobileEquipmentError -trimprefix=MMMobileEquipmentError
const (
	/* General errors */
	MmMobileEquipmentErrorPhoneFailure          MMMobileEquipmentError = 0   // Phone failure.
	MmMobileEquipmentErrorNoConnection          MMMobileEquipmentError = 1   // No connection to phone.
	MmMobileEquipmentErrorLinkReserved          MMMobileEquipmentError = 2   // Phone-adaptor link reserved.
	MmMobileEquipmentErrorNotAllowed            MMMobileEquipmentError = 3   // Operation not allowed.
	MmMobileEquipmentErrorNotSupported          MMMobileEquipmentError = 4   // Operation not supported.
	MmMobileEquipmentErrorPhSimPin              MMMobileEquipmentError = 5   // PH-SIM PIN required.
	MmMobileEquipmentErrorPhFsimPin             MMMobileEquipmentError = 6   // PH-FSIM PIN required.
	MmMobileEquipmentErrorPhFsimPuk             MMMobileEquipmentError = 7   // PH-FSIM PUK required.
	MmMobileEquipmentErrorSimNotInserted        MMMobileEquipmentError = 10  // SIM not inserted.
	MmMobileEquipmentErrorSimPin                MMMobileEquipmentError = 11  // SIM PIN required.
	MmMobileEquipmentErrorSimPuk                MMMobileEquipmentError = 12  // SIM PUK required.
	MmMobileEquipmentErrorSimFailure            MMMobileEquipmentError = 13  // SIM failure.
	MmMobileEquipmentErrorSimBusy               MMMobileEquipmentError = 14  // SIM busy.
	MmMobileEquipmentErrorSimWrong              MMMobileEquipmentError = 15  // SIM wrong.
	MmMobileEquipmentErrorIncorrectPassword     MMMobileEquipmentError = 16  // Incorrect password.
	MmMobileEquipmentErrorSimPin2               MMMobileEquipmentError = 17  // SIM PIN2 required.
	MmMobileEquipmentErrorSimPuk2               MMMobileEquipmentError = 18  // SIM PUK2 required.
	MmMobileEquipmentErrorMemoryFull            MMMobileEquipmentError = 20  // Memory full.
	MmMobileEquipmentErrorInvalidIndex          MMMobileEquipmentError = 21  // Invalid index.
	MmMobileEquipmentErrorNotFound              MMMobileEquipmentError = 22  // Not found.
	MmMobileEquipmentErrorMemoryFailure         MMMobileEquipmentError = 23  // Memory failure.
	MmMobileEquipmentErrorTextTooLong           MMMobileEquipmentError = 24  // Text string too long.
	MmMobileEquipmentErrorInvalidChars          MMMobileEquipmentError = 25  // Invalid characters in text string.
	MmMobileEquipmentErrorDialStringTooLong     MMMobileEquipmentError = 26  // Dial string too long.
	MmMobileEquipmentErrorDialStringInvalid     MMMobileEquipmentError = 27  // Invalid characters in dial string.
	MmMobileEquipmentErrorNoNetwork             MMMobileEquipmentError = 30  // No network service.
	MmMobileEquipmentErrorNetworkTimeout        MMMobileEquipmentError = 31  // Network timeout.
	MmMobileEquipmentErrorNetworkNotAllowed     MMMobileEquipmentError = 32  // Network not allowed - Emergency calls only.
	MmMobileEquipmentErrorNetworkPin            MMMobileEquipmentError = 40  // Network personalisation PIN required.
	MmMobileEquipmentErrorNetworkPuk            MMMobileEquipmentError = 41  // Network personalisation PUK required.
	MmMobileEquipmentErrorNetworkSubsetPin      MMMobileEquipmentError = 42  // Network subset personalisation PIN required.
	MmMobileEquipmentErrorNetworkSubsetPuk      MMMobileEquipmentError = 43  // Network subset personalisation PUK required.
	MmMobileEquipmentErrorServicePin            MMMobileEquipmentError = 44  // Service provider personalisation PIN required.
	MmMobileEquipmentErrorServicePuk            MMMobileEquipmentError = 45  // Service provider personalisation PUK required.
	MmMobileEquipmentErrorCorpPin               MMMobileEquipmentError = 46  // Corporate personalisation PIN required.
	MmMobileEquipmentErrorCorpPuk               MMMobileEquipmentError = 47  // Corporate personalisation PUK required.
	MmMobileEquipmentErrorHiddenKeyRequired     MMMobileEquipmentError = 48  // Hidden key required. Since: 1.8.
	MmMobileEquipmentErrorEapMethodNotSupported MMMobileEquipmentError = 49  // EAP method not supported. Since: 1.8.
	MmMobileEquipmentErrorIncorrectParameters   MMMobileEquipmentError = 50  // Incorrect parameters. Since: 1.8.
	MmMobileEquipmentErrorUnknown               MMMobileEquipmentError = 100 // Unknown.
	/* GPRS related errors */
	MmMobileEquipmentErrorGprsImsiUnknownInHlr                     MMMobileEquipmentError = 102 // IMSI unknown in HLR.
	MmMobileEquipmentErrorGprsIllegalMs                            MMMobileEquipmentError = 103 // IMSI unknown in VLR.
	MmMobileEquipmentErrorGprsImsiUnknownInVlr                     MMMobileEquipmentError = 104 // Illegal MS.
	MmMobileEquipmentErrorGprsIllegalMe                            MMMobileEquipmentError = 106 // Illegal ME.
	MmMobileEquipmentErrorGprsServiceNotAllowed                    MMMobileEquipmentError = 107 // GPRS service not allowed.
	MmMobileEquipmentErrorGprsAndNonGprsServicesNotAllowed         MMMobileEquipmentError = 108 // GPRS and non-GPRS services not allowed. Since: 1.8.
	MmMobileEquipmentErrorGprsPlmnNotAllowed                       MMMobileEquipmentError = 111 // PLMN not allowed.
	MmMobileEquipmentErrorGprsLocationNotAllowed                   MMMobileEquipmentError = 112 // Location area not allowed.
	MmMobileEquipmentErrorGprsRoamingNotAllowed                    MMMobileEquipmentError = 113 // Roaming not allowed in this location area.
	MmMobileEquipmentErrorGprsNoCellsInLocationArea                MMMobileEquipmentError = 115 // No cells in this location area.
	MmMobileEquipmentErrorGprsNetworkFailure                       MMMobileEquipmentError = 117 // Network failure.
	MmMobileEquipmentErrorGprsCongestion                           MMMobileEquipmentError = 122 // Congestion.
	MmMobileEquipmentErrorGprsNotAuthorizedForCsg                  MMMobileEquipmentError = 125 // GPRS not authorized for CSG. Since: 1.8.
	MmMobileEquipmentErrorGprsInsufficientResources                MMMobileEquipmentError = 126 // Insufficient resources. Since 1.4.
	MmMobileEquipmentErrorGprsMissingOrUnknownApn                  MMMobileEquipmentError = 127 // Missing or unknown APN. Since 1.4.
	MmMobileEquipmentErrorGprsUnknownPdpAddressOrType              MMMobileEquipmentError = 128 // Unknown PDP address or type. Since: 1.8.
	MmMobileEquipmentErrorGprsUserAuthenticationFailed             MMMobileEquipmentError = 129 // User authentication failed. Since 1.4.
	MmMobileEquipmentErrorGprsActivationRejectedByGgsnOrGw         MMMobileEquipmentError = 130 // Activation rejected by GGSN or gateway. Since: 1.8.
	MmMobileEquipmentErrorGprsActivationRejectedUnspecified        MMMobileEquipmentError = 131 // Activation rejected (reason unspecified). Since: 1.8.
	MmMobileEquipmentErrorGprsServiceOptionNotSupported            MMMobileEquipmentError = 132 // Service option not supported.
	MmMobileEquipmentErrorGprsServiceOptionNotSubscribed           MMMobileEquipmentError = 133 // Requested service option not subscribed.
	MmMobileEquipmentErrorGprsServiceOptionOutOfOrder              MMMobileEquipmentError = 134 // Service option temporarily out of order.
	MmMobileEquipmentErrorGprsFeatureNotSupported                  MMMobileEquipmentError = 140 // Feature not supported. Since: 1.8.
	MmMobileEquipmentErrorGprsSemanticErrorInTftOperation          MMMobileEquipmentError = 141 // Semantic error in TFT operation. Since: 1.8.
	MmMobileEquipmentErrorGprsSyntacticalErrorInTftOperation       MMMobileEquipmentError = 142 // Syntactical error in TFT operation. Since: 1.8.
	MmMobileEquipmentErrorGprsUnknownPdpContext                    MMMobileEquipmentError = 143 // Unknown PDP context. Since: 1.8.
	MmMobileEquipmentErrorGprsSemanticErrorsInPacketFilter         MMMobileEquipmentError = 144 // Semantic errors in packet filter. Since: 1.8.
	MmMobileEquipmentErrorGprsSyntacticalErrorInPacketFilter       MMMobileEquipmentError = 145 // Syntactical error in packet filter. Since: 1.8.
	MmMobileEquipmentErrorGprsPdpContextWithoutTftAlreadyActivated MMMobileEquipmentError = 146 // PDP context witout TFT already activated. Since: 1.8.
	MmMobileEquipmentErrorGprsUnknown                              MMMobileEquipmentError = 148 // Unspecified GPRS error.
	MmMobileEquipmentErrorGprsPdpAuthFailure                       MMMobileEquipmentError = 149 // PDP authentication failure.
	MmMobileEquipmentErrorGprsInvalidMobileClass                   MMMobileEquipmentError = 150 // Invalid mobile class.
	MmMobileEquipmentErrorGprsLastPdnDisconnectionNotAllowedLegacy MMMobileEquipmentError = 151 // Last PDN disconnection not allowed (legacy value defined before 3GPP Rel-11). Since: 1.14.
	MmMobileEquipmentErrorGprsLastPdnDisconnectionNotAllowed       MMMobileEquipmentError = 171 // Last PDN disconnection not allowed. Since: 1.8.
	MmMobileEquipmentErrorGprsSemanticallyIncorrectMessage         MMMobileEquipmentError = 172 // Semantically incorrect message. Since: 1.8.
	MmMobileEquipmentErrorGprsMandatoryIeError                     MMMobileEquipmentError = 173 // Mandatory IE error. Since: 1.8.
	MmMobileEquipmentErrorGprsIeNotImplemented                     MMMobileEquipmentError = 174 // IE not implemented. Since: 1.8.
	MmMobileEquipmentErrorGprsConditionalIeError                   MMMobileEquipmentError = 175 // Conditional IE error. Since: 1.8.
	MmMobileEquipmentErrorGprsUnspecifiedProtocolError             MMMobileEquipmentError = 176 // Unspecified protocol error. Since: 1.8.
	MmMobileEquipmentErrorGprsOperatorDeterminedBarring            MMMobileEquipmentError = 177 // Operator determined barring. Since: 1.8.
	MmMobileEquipmentErrorGprsMaximumNumberOfPdpContextsReached    MMMobileEquipmentError = 178 // Maximum number of PDP contexts reached. Since: 1.8.
	MmMobileEquipmentErrorGprsRequestedApnNotSupported             MMMobileEquipmentError = 179 // Requested APN not supported. Since: 1.8.
	MmMobileEquipmentErrorGprsRequestRejectedBcmViolation          MMMobileEquipmentError = 180 // Request rejected (BCM violation). Since: 1.8.

)

// MMConnectionError Connection errors that may be reported by ModemManager.
type MMConnectionError uint32

//go:generate stringer -type=MMConnectionError -trimprefix=MMConnectionError
const (
	MmConnectionErrorUnknown    MMConnectionError = 0 // Unknown connection error.
	MmConnectionErrorNoCarrier  MMConnectionError = 1 // No carrier.
	MmConnectionErrorNoDialtone MMConnectionError = 2 // No dialtone.
	MmConnectionErrorBusy       MMConnectionError = 3 // Busy.
	MmConnectionErrorNoAnswer   MMConnectionError = 4 // No answer.

)

// MMSerialError Serial errors that may be reported by ModemManager.
type MMSerialError uint32

//go:generate stringer -type=MMSerialError -trimprefix=MMSerialError
const (
	MmSerialErrorUnknown            MMSerialError = 0 // Unknown serial error.
	MmSerialErrorOpenFailed         MMSerialError = 1 // Could not open the serial device.
	MmSerialErrorSendFailed         MMSerialError = 2 // Could not write to the serial device.
	MmSerialErrorResponseTimeout    MMSerialError = 3 // A response was not received on time.
	MmSerialErrorOpenFailedNoDevice MMSerialError = 4 // Could not open the serial port, no device.
	MmSerialErrorFlashFailed        MMSerialError = 5 // Could not flash the device.
	MmSerialErrorNotOpen            MMSerialError = 6 // The serial port is not open.
	MmSerialErrorParseFailed        MMSerialError = 7 // The serial port specific parsing failed.
	MmSerialErrorFrameNotFound      MMSerialError = 8 // The serial port reported that the frame marker wasn't found (e.g. for QCDM). Since 1.6.

)

// MMMessageError Enumeration of message errors, as defined in 3GPP TS 27.005 version 10 section 3.2.5.
type MMMessageError uint32

//go:generate stringer -type=MMMessageError -trimprefix=MMMessageError
const (
	/* 0 -> 127 per 3GPP TS 24.011 [6] clause E.2 */
	/* 128 -> 255 per 3GPP TS 23.040 [3] clause 9.2.3.22 */
	MmMessageErrorMeFailure            MMMessageError = 300 // ME failure.
	MmMessageErrorSmsServiceReserved   MMMessageError = 301 // SMS service reserved.
	MmMessageErrorNotAllowed           MMMessageError = 302 // Operation not allowed.
	MmMessageErrorNotSupported         MMMessageError = 303 // Operation not supported.
	MmMessageErrorInvalidPduParameter  MMMessageError = 304 // Invalid PDU mode parameter.
	MmMessageErrorInvalidTextParameter MMMessageError = 305 // Invalid text mode parameter.
	MmMessageErrorSimNotInserted       MMMessageError = 310 // SIM not inserted.
	MmMessageErrorSimPin               MMMessageError = 311 // SIM PIN required.
	MmMessageErrorPhSimPin             MMMessageError = 312 // PH-SIM PIN required.
	MmMessageErrorSimFailure           MMMessageError = 313 // SIM failure.
	MmMessageErrorSimBusy              MMMessageError = 314 // SIM busy.
	MmMessageErrorSimWrong             MMMessageError = 315 // SIM wrong.
	MmMessageErrorSimPuk               MMMessageError = 316 // SIM PUK required.
	MmMessageErrorSimPin2              MMMessageError = 317 // SIM PIN2 required.
	MmMessageErrorSimPuk2              MMMessageError = 318 // SIM PUK2 required.
	MmMessageErrorMemoryFailure        MMMessageError = 320 // Memory failure.
	MmMessageErrorInvalidIndex         MMMessageError = 321 // Invalid index.
	MmMessageErrorMemoryFull           MMMessageError = 322 // Memory full.
	MmMessageErrorSmscAddressUnknown   MMMessageError = 330 // SMSC address unknown.
	MmMessageErrorNoNetwork            MMMessageError = 331 // No network.
	MmMessageErrorNetworkTimeout       MMMessageError = 332 // Network timeout.
	MmMessageErrorNoCnmaAckExpected    MMMessageError = 340 // No CNMA Acknowledgement expected.
	MmMessageErrorUnknown              MMMessageError = 500 // Unknown error.

)

// MMCdmaActivationError CDMA Activation errors
type MMCdmaActivationError uint32

//go:generate stringer -type=MMCdmaActivationError -trimprefix=MmCdmaActivationError
const (
	MmCdmaActivationErrorNone                         MMCdmaActivationError = 0 // No error.
	MmCdmaActivationErrorUnknown                      MMCdmaActivationError = 1 // An error occurred.
	MmCdmaActivationErrorRoaming                      MMCdmaActivationError = 2 // Device cannot activate while roaming.
	MmCdmaActivationErrorWrongRadioInterface          MMCdmaActivationError = 3 // Device cannot activate on this network type (eg EVDO vs 1xRTT).
	MmCdmaActivationErrorCouldNotConnect              MMCdmaActivationError = 4 // Device could not connect to the network for activation.
	MmCdmaActivationErrorSecurityAuthenticationFailed MMCdmaActivationError = 5 // Device could not authenticate to the network for activation.
	MmCdmaActivationErrorProvisioningFailed           MMCdmaActivationError = 6 // Later stages of device provisioning failed.
	MmCdmaActivationErrorNoSignal                     MMCdmaActivationError = 7 // No signal available.
	MmCdmaActivationErrorTimedOut                     MMCdmaActivationError = 8 // Activation timed out.
	MmCdmaActivationErrorStartFailed                  MMCdmaActivationError = 9 // API call for initial activation failed.

)

// MMSignalPropertyType SignalProperty Type
type MMSignalPropertyType uint32

//go:generate stringer -type=MMSignalPropertyType -trimprefix=MMSignalPropertyType
const (
	MMSignalPropertyTypeCdma MMSignalPropertyType = 0 // Signal Type Cdma.
	MMSignalPropertyTypeEvdo MMSignalPropertyType = 1 // Signal Type Evdo.
	MMSignalPropertyTypeGsm  MMSignalPropertyType = 2 // Signal Type Gsm.
	MMSignalPropertyTypeUmts MMSignalPropertyType = 3 // Signal Type Umts.
	MMSignalPropertyTypeLte  MMSignalPropertyType = 4 // Signal Type Lte.

)

// ref https://gitlab.freedesktop.org/mobile-broadband/ModemManager/-/blob/master/include/ModemManager-enums.h
