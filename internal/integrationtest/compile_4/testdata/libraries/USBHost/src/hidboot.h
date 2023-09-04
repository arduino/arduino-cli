/* Copyright (C) 2011 Circuits At Home, LTD. All rights reserved.

This software may be distributed and modified under the terms of the GNU
General Public License version 2 (GPL2) as published by the Free Software
Foundation and appearing in the file GPL2.TXT included in the packaging of
this file. Please note that GPL2 Section 2[b] requires that all works based
on this software must also be made publicly available under the terms of
the GPL2 ("Copyleft").

Contact information
-------------------

Circuits At Home, LTD
Web      :  http://www.circuitsathome.com
e-mail   :  support@circuitsathome.com
*/

#ifndef HIDBOOT_H_INCLUDED
#define HIDBOOT_H_INCLUDED

#include <stdint.h>
#include "Usb.h"
#include "hid.h"
#include "Arduino.h"

#define UHS_HID_BOOT_KEY_ZERO           0x27
#define UHS_HID_BOOT_KEY_ENTER          0x28
#define UHS_HID_BOOT_KEY_ESCAPE         0x29
#define UHS_HID_BOOT_KEY_DELETE         0x2a    // Backspace
#define UHS_HID_BOOT_KEY_DELETE_FORWARD	0x4C	// Delete
#define UHS_HID_BOOT_KEY_TAB            0x2b
#define UHS_HID_BOOT_KEY_SPACE          0x2c
#define UHS_HID_BOOT_KEY_CAPS_LOCK      0x39
#define UHS_HID_BOOT_KEY_SCROLL_LOCK    0x47
#define UHS_HID_BOOT_KEY_NUM_LOCK       0x53
#define UHS_HID_BOOT_KEY_ZERO2          0x62
#define UHS_HID_BOOT_KEY_PERIOD         0x63

// Don't worry, GCC will optimize the result to a final value.
#define bitsEndpoints(p) ((((p) & HID_PROTOCOL_KEYBOARD)? 2 : 0) | (((p) & HID_PROTOCOL_MOUSE)? 1 : 0))
#define totalEndpoints(p) ((bitsEndpoints(p) == 3) ? 3 : 2)
#define epMUL(p) ((((p) & HID_PROTOCOL_KEYBOARD)? 1 : 0) + (((p) & HID_PROTOCOL_MOUSE)? 1 : 0))

// Already defined in hid.h
// #define HID_MAX_HID_CLASS_DESCRIPTORS 5

struct MOUSEINFO {

        struct {
                uint8_t bmLeftButton : 1;
                uint8_t bmRightButton : 1;
                uint8_t bmMiddleButton : 1;
                uint8_t bmDummy : 5;
	};
	int8_t			dX;
	int8_t			dY;
};

class MouseReportParser : public HIDReportParser {

        union {
                MOUSEINFO mouseInfo;
                uint8_t bInfo[sizeof (MOUSEINFO)];
        } prevState;

public:
	virtual void Parse(HID *hid, uint32_t is_rpt_id, uint32_t len, uint8_t *buf);

protected:

        virtual void OnMouseMove(MOUSEINFO *) {
        };

        virtual void OnLeftButtonUp(MOUSEINFO *) {
        };

        virtual void OnLeftButtonDown(MOUSEINFO *) {
        };

        virtual void OnRightButtonUp(MOUSEINFO *) {
        };

        virtual void OnRightButtonDown(MOUSEINFO *) {
        };

        virtual void OnMiddleButtonUp(MOUSEINFO *) {
        };

        virtual void OnMiddleButtonDown(MOUSEINFO *) {
        };
};

struct MODIFIERKEYS {
	uint8_t		bmLeftCtrl		: 1;
	uint8_t		bmLeftShift		: 1;
	uint8_t		bmLeftAlt		: 1;
	uint8_t		bmLeftGUI		: 1;
	uint8_t		bmRightCtrl		: 1;
	uint8_t		bmRightShift	: 1;
	uint8_t		bmRightAlt		: 1;
	uint8_t		bmRightGUI		: 1;
};

struct KBDINFO {

        struct {
		uint8_t		bmLeftCtrl		: 1;
		uint8_t		bmLeftShift		: 1;
		uint8_t		bmLeftAlt		: 1;
		uint8_t		bmLeftGUI		: 1;
		uint8_t		bmRightCtrl		: 1;
		uint8_t		bmRightShift	: 1;
		uint8_t		bmRightAlt		: 1;
		uint8_t		bmRightGUI		: 1;
	};
	uint8_t			bReserved;
	uint8_t			Keys[6];
};

struct KBDLEDS {
	uint8_t		bmNumLock		: 1;
	uint8_t		bmCapsLock		: 1;
	uint8_t		bmScrollLock	: 1;
	uint8_t		bmCompose		: 1;
	uint8_t		bmKana			: 1;
	uint8_t		bmReserved		: 3;
};

class KeyboardReportParser : public HIDReportParser {
        static const uint8_t numKeys[10];
        static const uint8_t symKeysUp[12];
        static const uint8_t symKeysLo[12];
        static const uint8_t padKeys[5];

protected:

        union {
                KBDINFO kbdInfo;
                uint8_t bInfo[sizeof (KBDINFO)];
        } prevState;

        union {
		KBDLEDS		kbdLeds;
		uint8_t		bLeds;
	}	kbdLockingKeys;

	uint8_t OemToAscii(uint8_t mod, uint8_t key);

public:

        KeyboardReportParser() {
                kbdLockingKeys.bLeds = 0;
        };

	virtual void Parse(HID *hid, uint32_t is_rpt_id, uint32_t len, uint8_t *buf);

protected:
	virtual uint8_t HandleLockingKeys(HID* hid, uint8_t key);

        virtual void OnControlKeysChanged(uint8_t /* before */, uint8_t /* after */) {
        };

        virtual void OnKeyDown(uint8_t /* mod */, uint8_t /* key */) {
        };

        virtual void OnKeyUp(uint8_t /* mod */, uint8_t /* key */) {
        };

        virtual const uint8_t *getNumKeys() {
                return numKeys;
        };

        virtual const uint8_t *getSymKeysUp() {
                return symKeysUp;
        };

        virtual const uint8_t *getSymKeysLo() {
                return symKeysLo;
        };

        virtual const uint8_t *getPadKeys() {
                return padKeys;
        };
};

template <const uint8_t BOOT_PROTOCOL>
class HIDBoot : public HID //public USBDeviceConfig, public UsbConfigXtracter
{
        EpInfo epInfo[totalEndpoints(BOOT_PROTOCOL)];
        HIDReportParser *pRptParser[epMUL(BOOT_PROTOCOL)];

        uint32_t bConfNum; // configuration number
        uint32_t bIfaceNum; // Interface Number
        uint32_t bNumIface; // number of interfaces in the configuration
        uint32_t bNumEP; // total number of EP in the configuration

        uint32_t qNextPollTime;			// next poll time
        uint32_t bPollEnable;			// poll enable flag
        uint32_t bInterval; // largest interval

	void Initialize();

	virtual HIDReportParser* GetReportParser(uint32_t id) { 
                return pRptParser[id];
        };

public:
	HIDBoot(USBHost *p);

	virtual uint32_t SetReportParser(uint32_t id, HIDReportParser *prs) {
                pRptParser[id] = prs;
                return true;
        };

	// USBDeviceConfig implementation
	virtual uint32_t Init(uint32_t parent, uint32_t port, uint32_t lowspeed);
	virtual uint32_t Release();
	virtual uint32_t Poll();

	virtual uint32_t GetAddress() {
                return bAddress;
        };

	// UsbConfigXtracter implementation
	virtual void EndpointXtract(uint32_t conf, uint32_t iface, uint32_t alt, uint32_t proto, const USB_ENDPOINT_DESCRIPTOR *ep);
};

template <const uint8_t BOOT_PROTOCOL>
HIDBoot<BOOT_PROTOCOL>::HIDBoot(USBHost *p) :
		HID(p),
		qNextPollTime(0),
		bPollEnable(false) {
	Initialize();

        for(int i = 0; i < epMUL(BOOT_PROTOCOL); i++) {
                pRptParser[i] = NULL;
        }
	if(pUsb)
		pUsb->RegisterDeviceClass(this);
}

template <const uint8_t BOOT_PROTOCOL>
void HIDBoot<BOOT_PROTOCOL>::Initialize() {
        for(int i = 0; i < totalEndpoints(BOOT_PROTOCOL); i++) {
                epInfo[i].epAddr = 0;
		epInfo[i].maxPktSize	= (i) ? 0 : 8;
		epInfo[i].bmSndToggle   = 0;
		epInfo[i].bmRcvToggle   = 0;
		epInfo[i].bmNakPower	= (i) ? USB_NAK_NOWAIT : USB_NAK_MAX_POWER;
	}
	bNumEP		= 1;
	bNumIface	= 0;
	bConfNum	= 0;
}

template <const uint8_t BOOT_PROTOCOL>
uint32_t HIDBoot<BOOT_PROTOCOL>::Init(uint32_t parent, uint32_t port, uint32_t lowspeed) {
        const uint8_t constBufSize = sizeof (USB_DEVICE_DESCRIPTOR);

	uint8_t		buf[constBufSize];
	uint32_t	rcode;
	UsbDeviceDefinition	*p = NULL;
	EpInfo		*oldep_ptr = NULL;
	uint32_t	len = 0;
    //uint16_t cd_len = 0;

	uint32_t num_of_conf; // number of configurations
    //uint32_t num_of_intf; // number of interfaces

	// Get memory address of USB device address pool
	AddressPool	&addrPool = pUsb->GetAddressPool();

	USBTRACE("BM Init\r\n");
	//USBTRACE2("totalEndpoints:", (uint8_t) (totalEndpoints(BOOT_PROTOCOL)));
	//USBTRACE2("epMUL:", epMUL(BOOT_PROTOCOL));

    // Check if address has already been assigned to an instance
	if(bAddress)
		return USB_ERROR_CLASS_INSTANCE_ALREADY_IN_USE;

    bInterval = 0;
	// Get pointer to pseudo device with address 0 assigned
	p = addrPool.GetUsbDevicePtr(0);

	if(!p)
		return USB_ERROR_ADDRESS_NOT_FOUND_IN_POOL;

	if(!p->epinfo) {
                USBTRACE("epinfo\r\n");
		return USB_ERROR_EPINFO_IS_NULL;
	}

	// Save old pointer to EP_RECORD of address 0
	oldep_ptr = p->epinfo;

	// Temporary assign new pointer to epInfo to p->epinfo in order to avoid toggle inconsistence
	p->epinfo = epInfo;

	p->lowspeed = lowspeed;
	// Get device descriptor  GET_DESCRIPTOR
	rcode = pUsb->getDevDescr(0, 0, /*8*/sizeof(USB_DEVICE_DESCRIPTOR), (uint8_t*)buf);

	if(!rcode)
		len = (buf[0] > constBufSize) ? constBufSize : buf[0];

    if(rcode) {
		// Restore p->epinfo
		p->epinfo = oldep_ptr;

		goto FailGetDevDescr;
    }


	// Restore p->epinfo
	p->epinfo = oldep_ptr;


	// Allocate new address according to device class
	bAddress = addrPool.AllocAddress(parent, false, port);
	if(!bAddress)
		return USB_ERROR_OUT_OF_ADDRESS_SPACE_IN_POOL;

	// Extract Max Packet Size from the device descriptor
	epInfo[0].maxPktSize = (uint8_t)((USB_DEVICE_DESCRIPTOR*)buf)->bMaxPacketSize0;

	// Assign new address to the device  SET_ADDRESS

	rcode = pUsb->setAddr(0, 0, bAddress);
	if(rcode) {
		p->lowspeed = false;
		addrPool.FreeAddress(bAddress);
		bAddress = 0;
		USBTRACE2("setAddr:", rcode);
		return rcode;
	}
    //delay(2); //per USB 2.0 sect.9.2.6.3

	// Save old pointer to EP_RECORD of address 0
	oldep_ptr = p->epinfo;
	// Temporary assign new pointer to epInfo to p->epinfo in order to avoid toggle inconsistence
	p->epinfo = epInfo;
	// Get device descriptor  GET_DESCRIPTOR
	rcode = pUsb->getDevDescr(bAddress, 0, sizeof(USB_DEVICE_DESCRIPTOR), (uint8_t*)buf);
	// Restore p->epinfo
	p->epinfo = oldep_ptr;
	if (rcode)
	{
		goto FailGetDevDescr;
	}

	TRACE_USBHOST(printf("HIDBoot::Init : device address is now %lu\r\n", bAddress);)

	p->lowspeed = false;

	//get pointer to assigned address record
	p = addrPool.GetUsbDevicePtr(bAddress);

	if(!p)
		return USB_ERROR_ADDRESS_NOT_FOUND_IN_POOL;

	p->lowspeed = lowspeed;

	if(len)
		rcode = pUsb->getDevDescr(bAddress, 0, len, (uint8_t*)buf);

	if(rcode)
		goto FailGetDevDescr;

	num_of_conf = ((USB_DEVICE_DESCRIPTOR*)buf)->bNumConfigurations;

        USBTRACE2("NC:", num_of_conf);

        // GCC will optimize unused stuff away.
        if((BOOT_PROTOCOL & (HID_PROTOCOL_KEYBOARD | HID_PROTOCOL_MOUSE)) == (HID_PROTOCOL_KEYBOARD | HID_PROTOCOL_MOUSE)) {
                USBTRACE("HID_PROTOCOL_KEYBOARD AND MOUSE\r\n");
                ConfigDescParser<
                        USB_CLASS_HID,
                        HID_BOOT_INTF_SUBCLASS,
                        HID_PROTOCOL_KEYBOARD | HID_PROTOCOL_MOUSE,
                        CP_MASK_COMPARE_ALL > confDescrParser(this);
                confDescrParser.SetOR(); // Use the OR variant.
                for(uint32_t i = 0; i < num_of_conf; i++) {
                        pUsb->getConfDescr(bAddress, 0, i, &confDescrParser);
                        if(bNumEP == (uint8_t)(totalEndpoints(BOOT_PROTOCOL)))
                                break;
                }
        } else {
                // GCC will optimize unused stuff away.
                if(BOOT_PROTOCOL & HID_PROTOCOL_KEYBOARD) {
                        USBTRACE("HID_PROTOCOL_KEYBOARD\r\n");
                        for(uint32_t i = 0; i < num_of_conf; i++) {
                                ConfigDescParser<
                                        USB_CLASS_HID,
                                        HID_BOOT_INTF_SUBCLASS,
                                        HID_PROTOCOL_KEYBOARD,
                                        CP_MASK_COMPARE_ALL> confDescrParserA(this);

                                pUsb->getConfDescr(bAddress, 0, i, &confDescrParserA);
                                if(bNumEP == (uint8_t)(totalEndpoints(BOOT_PROTOCOL)))
                                        break;
                        }
                }

                // GCC will optimize unused stuff away.
                if(BOOT_PROTOCOL & HID_PROTOCOL_MOUSE) {
                        USBTRACE("HID_PROTOCOL_MOUSE\r\n");
                        for(uint32_t i = 0; i < num_of_conf; i++) {
                                ConfigDescParser<
                                        USB_CLASS_HID,
                                        HID_BOOT_INTF_SUBCLASS,
                                        HID_PROTOCOL_MOUSE,
                                        CP_MASK_COMPARE_ALL> confDescrParserB(this);

                                pUsb->getConfDescr(bAddress, 0, i, &confDescrParserB);
                                if(bNumEP == ((uint8_t)(totalEndpoints(BOOT_PROTOCOL))))
                                        break;

                        }
                }
        }
        USBTRACE2("bNumEP:", bNumEP);

        if(bNumEP != (uint8_t)(totalEndpoints(BOOT_PROTOCOL))) {
                rcode = USB_DEV_CONFIG_ERROR_DEVICE_NOT_SUPPORTED;
                goto Fail;
        }

	// Assign epInfo to epinfo pointer
	rcode = pUsb->setEpInfoEntry(bAddress, bNumEP, epInfo);
        //USBTRACE2("setEpInfoEntry returned ", rcode);
        USBTRACE2("Cnf:", bConfNum);

        delay(1000);

	// Set Configuration Value
	rcode = pUsb->setConf(bAddress, 0, bConfNum);

	if(rcode)
		goto FailSetConfDescr;

        delay(1000);

        USBTRACE2("bIfaceNum:", bIfaceNum);
        USBTRACE2("bNumIface:", bNumIface);

        // Yes, mouse wants SetProtocol and SetIdle too!
        for(uint32_t i = 0; i < epMUL(BOOT_PROTOCOL); i++) {
                USBTRACE2("\r\nInterface:", i);
                rcode = SetProtocol(i, HID_BOOT_PROTOCOL);
                if(rcode) goto FailSetProtocol;
                USBTRACE2("PROTOCOL SET HID_BOOT rcode:", rcode);
                rcode = SetIdle(i, 0, 0);
                USBTRACE2("SET_IDLE rcode:", rcode);
                // if(rcode) goto FailSetIdle; This can fail.
                // Get the RPIPE and just throw it away.
                SinkParser <USBReadParser, uint32_t, uint32_t> sink;
                rcode = GetReportDescr(i, (USBReadParser *)&sink);
                USBTRACE2("RPIPE rcode:", rcode);
        }

        // Get RPIPE and throw it away.

        if(BOOT_PROTOCOL & HID_PROTOCOL_KEYBOARD) {
                // Wake keyboard interface by twinkling up to 5 LEDs that are in the spec.
                // kana, compose, scroll, caps, num
                rcode = 0x20; // Reuse rcode.
                while(rcode) {
                        rcode >>= 1;
                        // Ignore any error returned, we don't care if LED is not supported
                        SetReport(0, 0, 2, 0, 1, (uint8_t*)&rcode); // Eventually becomes zero (All off)
                        delay(25);
                }
        }
        USBTRACE("BM configured\r\n");

	bPollEnable = true;
	return 0;

FailGetDevDescr:
#ifdef DEBUG_USB_HOST
        NotifyFailGetDevDescr();
        goto Fail;
#endif

        //FailSetDevTblEntry:
        //#ifdef DEBUG_USB_HOST
        //        NotifyFailSetDevTblEntry();
        //        goto Fail;
        //#endif

        //FailGetConfDescr:
        //#ifdef DEBUG_USB_HOST
        //        NotifyFailGetConfDescr();
        //        goto Fail;
        //#endif

FailSetConfDescr:
#ifdef DEBUG_USB_HOST
        NotifyFailSetConfDescr();
        goto Fail;
#endif

FailSetProtocol:
#ifdef DEBUG_USB_HOST
        USBTRACE("SetProto:");
        goto Fail;
#endif

        //FailSetIdle:
        //#ifdef DEBUG_USB_HOST
        //        USBTRACE("SetIdle:");
        //#endif

Fail:
#ifdef DEBUG_USB_HOST
        NotifyFail(rcode);
#endif
        Release();

        return rcode;
}

template <const uint8_t BOOT_PROTOCOL>
void HIDBoot<BOOT_PROTOCOL>::EndpointXtract(uint32_t conf, uint32_t iface, uint32_t /* alt */, uint32_t /* proto */, const USB_ENDPOINT_DESCRIPTOR *pep) {
	// If the first configuration satisfies, the others are not considered.
        //if(bNumEP > 1 && conf != bConfNum)
        if(bNumEP == totalEndpoints(BOOT_PROTOCOL))
		return;

	bConfNum = conf;
	bIfaceNum = iface;

	if((pep->bmAttributes & 0x03) == 3 && (pep->bEndpointAddress & 0x80) == 0x80) {
                if(pep->bInterval > bInterval) bInterval = pep->bInterval;

		// Fill in the endpoint info structure
		epInfo[bNumEP].epAddr		= (pep->bEndpointAddress & 0x0F);
		epInfo[bNumEP].maxPktSize	= (uint8_t)pep->wMaxPacketSize;
		epInfo[bNumEP].bmSndToggle = 0;
		epInfo[bNumEP].bmRcvToggle = 0;
                epInfo[bNumEP].bmNakPower = USB_NAK_NOWAIT;
		bNumEP++;

	}
}

template <const uint8_t BOOT_PROTOCOL>
uint32_t HIDBoot<BOOT_PROTOCOL>::Release() {
	pUsb->GetAddressPool().FreeAddress(bAddress);

	bConfNum			= 0;
	bIfaceNum			= 0;
	bNumEP				= 1;
	bAddress			= 0;
	qNextPollTime		= 0;
	bPollEnable			= false;

	return 0;
}

template <const uint8_t BOOT_PROTOCOL>
uint32_t HIDBoot<BOOT_PROTOCOL>::Poll() {
	uint32_t rcode = 0;

        if(bPollEnable && ((long)(millis() - qNextPollTime) >= 0L)) {

                // To-do: optimize manually, using the for loop only if needed.
                for(int i = 0; i < epMUL(BOOT_PROTOCOL); i++) {
                        const uint16_t const_buff_len = 16;
                        uint8_t buf[const_buff_len];

                        USBTRACE3("(hidboot.h) i=", i, 0x81);
                        USBTRACE3("(hidboot.h) epInfo[epInterruptInIndex + i].epAddr=", epInfo[epInterruptInIndex + i].epAddr, 0x81);
                        USBTRACE3("(hidboot.h) epInfo[epInterruptInIndex + i].maxPktSize=", epInfo[epInterruptInIndex + i].maxPktSize, 0x81);
                        uint16_t read = (uint16_t)epInfo[epInterruptInIndex + i].maxPktSize;
                        UHD_Pipe_Alloc(bAddress, epInfo[epInterruptInIndex + i].epAddr, USB_HOST_PTYPE_BULK, USB_EP_DIR_IN, epInfo[epInterruptInIndex + i].maxPktSize, 0, USB_HOST_NB_BK_1);
                        rcode = pUsb->inTransfer(bAddress, epInfo[epInterruptInIndex + i].epAddr, (uint8_t*)&read, buf);
                        // SOME buggy dongles report extra keys (like sleep) using a 2 byte packet on the wrong endpoint.
                        // Since keyboard and mice must report at least 3 bytes, we ignore the extra data.
                        if(!rcode && read > 2) {
                                if(pRptParser[i])
                                        pRptParser[i]->Parse((HID*)this, 0, (uint8_t)read, buf);
#ifdef DEBUG_USB_HOST
                                // We really don't care about errors and anomalies unless we are debugging.
                        } else {
                                if(rcode != USB_ERRORFLOW) {
                                        USBTRACE3("(hidboot.h) Poll:", rcode, 0x81);
                                }
                                if(!rcode && read) {
                                        USBTRACE3("(hidboot.h) Strange read count: ", read, 0x80);
                                        USBTRACE3("(hidboot.h) Interface:", i, 0x80);
                                }
                        }

                        if(!rcode && read && (UsbDEBUGlvl > 0x7f)) {
                                for(uint8_t i = 0; i < read; i++) {
                                        PrintHex<uint8_t > (buf[i], 0x80);
                                        USBTRACE1(" ", 0x80);
                                }
                                if(read)
                                        USBTRACE1("\r\n", 0x80);
#endif
                        }

                }
                qNextPollTime = millis() + bInterval;
        }
        return rcode;
}

#endif /* HIDBOOT_H_INCLUDED */
