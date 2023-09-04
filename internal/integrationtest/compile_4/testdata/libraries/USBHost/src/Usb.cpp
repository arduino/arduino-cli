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
/* USB functions */

#include <stdio.h>
#include "Arduino.h"
#include "Usb.h"


//#ifdef ARDUINO_SAMD_ZERO

static uint32_t usb_error = 0;
static uint32_t usb_task_state = USB_DETACHED_SUBSTATE_INITIALIZE;

/* constructor */
USBHost::USBHost() : bmHubPre(0) {
	// Set up state machine
	usb_task_state = USB_DETACHED_SUBSTATE_INITIALIZE; //set up state machine
}

/* Initialize data structures */
uint32_t USBHost::Init() {
	//devConfigIndex	= 0;
	// Init host stack
	bmHubPre		= 0;
	UHD_Init();
	return 0;
}

uint32_t USBHost::getUsbTaskState(void) {
    return (usb_task_state);
}

void USBHost::setUsbTaskState(uint32_t state) {
    usb_task_state = state;
}

uint32_t USBHost::getUsbErrorCode(void) {
     return (usb_error);
}
 
EpInfo* USBHost::getEpInfoEntry(uint32_t addr, uint32_t ep) {
	UsbDeviceDefinition *p = addrPool.GetUsbDevicePtr(addr);

	if(!p || !p->epinfo)
		return NULL;

	EpInfo *pep = p->epinfo;

	for (uint32_t i = 0; i < p->epcount; i++) {
		if(pep->epAddr == ep)
			return pep;

		pep++;
	}
	return NULL;
}

/* set device table entry */

/* each device is different and has different number of endpoints. This function plugs endpoint record structure, defined in application, to devtable */
uint32_t USBHost::setEpInfoEntry(uint32_t addr, uint32_t epcount, EpInfo* eprecord_ptr) {
	if (!eprecord_ptr)
		return USB_ERROR_INVALID_ARGUMENT;

	UsbDeviceDefinition *p = addrPool.GetUsbDevicePtr(addr);

	if(!p)
		return USB_ERROR_ADDRESS_NOT_FOUND_IN_POOL;

        p->address.devAddress = addr;
	p->epinfo	= eprecord_ptr;
	p->epcount	= epcount;

	return 0;
}

uint32_t USBHost::SetPipeAddress(uint32_t addr, uint32_t ep, EpInfo **ppep, uint32_t &nak_limit) {
	UsbDeviceDefinition *p = addrPool.GetUsbDevicePtr(addr);

	if(!p)
		return USB_ERROR_ADDRESS_NOT_FOUND_IN_POOL;

 	if(!p->epinfo)
		return USB_ERROR_EPINFO_IS_NULL;

	*ppep = getEpInfoEntry(addr, ep);

	if(!*ppep)
		return USB_ERROR_EP_NOT_FOUND_IN_TBL;

	nak_limit = (0x0001UL << (((*ppep)->bmNakPower > USB_NAK_MAX_POWER ) ? USB_NAK_MAX_POWER : (*ppep)->bmNakPower));
	nak_limit--;
	TRACE_USBHOST(printf("     => SetPipeAddress deviceEP=%lu configued as hostPIPE=%lu sending to address=%lu\r\n", ep, (*ppep)->epAddr, addr);)
	/*
          USBTRACE2("\r\nAddress: ", addr);
          USBTRACE2(" EP: ", ep);
          USBTRACE2(" NAK Power: ",(*ppep)->bmNakPower);
          USBTRACE2(" NAK Limit: ", nak_limit);
          USBTRACE("\r\n");
         */

	// CTRL_PIPE.PDADDR: usb_pipe_table[pipe_num].HostDescBank[0].CTRL_PIPE.bit.PDADDR = addr
	uhd_configure_address((*ppep)->epAddr, addr); 	// Set peripheral address

	return 0;
}

/* Control transfer. Sets address, endpoint, fills control packet with necessary data, dispatches control packet, and initiates bulk IN transfer,   */
/* depending on request. Actual requests are defined as inlines                                                                                      */
/* return codes:                */
/* 00       =   success         */
/* 01-0f    =   non-zero HRSLT  */
uint32_t USBHost::ctrlReq(uint32_t addr, uint32_t ep, uint8_t bmReqType, uint8_t bRequest, uint8_t wValLo, uint8_t wValHi,
                          uint16_t wInd, uint16_t total, uint32_t nbytes, uint8_t* dataptr, USBReadParser *p) {
	
	uint32_t direction = 0; // Request direction, IN or OUT
	uint32_t rcode;
	SETUP_PKT setup_pkt;

	EpInfo *pep = NULL;
	uint32_t nak_limit = 0;

	TRACE_USBHOST(printf("    => ctrlReq\r\n");)

	rcode = SetPipeAddress(addr, ep, &pep, nak_limit);
	if(rcode)
		return rcode;

	// Allocate Pipe0 with default 64 bytes size if not already initialized
	rcode = UHD_Pipe0_Alloc(0, 64);
	if (rcode)
	{
		TRACE_USBHOST(printf("/!\\ USBHost::ctrlReq : EP0 allocation error: %lu\r\n", rcode);)
		//USBTRACE2("\n\rUSBHost::ctrlReq : EP0 allocation error: ", rcode");
		return rcode;
	}

	// Determine request direction
	direction = ((bmReqType & 0x80 ) > 0);

	/* fill in setup packet */
    setup_pkt.ReqType_u.bmRequestType	= bmReqType;
    setup_pkt.bRequest					= bRequest;
    setup_pkt.wVal_u.wValueLo			= wValLo;
    setup_pkt.wVal_u.wValueHi			= wValHi;
    setup_pkt.wIndex					= wInd;
    setup_pkt.wLength					= total;

	UHD_Pipe_Write(pep->epAddr, sizeof(setup_pkt), (uint8_t *)&setup_pkt); //transfer to setup packet FIFO

	rcode = dispatchPkt(tokSETUP, ep, nak_limit); // Dispatch packet

	if (rcode) //return HRSLT if not zero
		return ( rcode);

	if (dataptr != NULL) //data stage, if present
	{
		if (direction) // IN transfer
		{
			uint32_t left = total;
			TRACE_USBHOST(printf("    => ctrlData IN\r\n");)

			pep->bmRcvToggle = 1; //bmRCVTOG1;

			// Bytes read into buffer
			uint32_t read = nbytes;

			rcode = InTransfer(pep, nak_limit, (uint8_t*)&read, dataptr);

			if((rcode&USB_ERROR_DATATOGGLE) == USB_ERROR_DATATOGGLE) {
						// yes, we flip it wrong here so that next time it is actually correct!
						//pep->bmRcvToggle = (regRd(rHRSL) & bmSNDTOGRD) ? 0 : 1;
						pep->bmRcvToggle = USB_HOST_DTGL(pep->epAddr);
						//continue;
			}

			if(rcode) {
				//USBTRACE2("\n\rUSBHost::ctrlReq : in transfer: ", rcode");
				return rcode;
			}
			// Invoke callback function if inTransfer completed successfully and callback function pointer is specified
			if(!rcode && p)
				((USBReadParser*)p)->Parse(read, dataptr, total - left);
		}
		else // OUT transfer
		{			
			pep->bmSndToggle = 1; //bmSNDTOG1;
			rcode = OutTransfer(pep, nak_limit, nbytes, dataptr);
		}
		if(rcode) //return error
			return (rcode);
	}
	
	// Status stage
	UHD_Pipe_CountZero(pep->epAddr);
	USB->HOST.HostPipe[pep->epAddr].PSTATUSSET.reg = USB_HOST_PSTATUSSET_DTGL;
	return dispatchPkt((direction) ? tokOUTHS : tokINHS, pep->epAddr, nak_limit); //GET if direction
}

/* IN transfer to arbitrary endpoint. Assumes PERADDR is set. Handles multiple packets if necessary. Transfers 'nbytes' bytes. */
/* Keep sending INs and writes data to memory area pointed by 'data'                                                           */

/* rcode 0 if no errors. rcode 01-0f is relayed from dispatchPkt(). Rcode f0 means RCVDAVIRQ error,
            fe USB xfer timeout */
uint32_t USBHost::inTransfer(uint32_t addr, uint32_t ep, uint8_t *nbytesptr, uint8_t* data) {
	EpInfo *pep = NULL;
	uint32_t nak_limit = 0;

	uint32_t rcode = SetPipeAddress(addr, ep, &pep, nak_limit);

        if(rcode) {
                USBTRACE3("(USB::InTransfer) SetAddress Failed ", rcode, 0x81);
                USBTRACE3("(USB::InTransfer) addr requested ", addr, 0x81);
                USBTRACE3("(USB::InTransfer) ep requested ", ep, 0x81);
                return rcode;
        }
	return InTransfer(pep, nak_limit, nbytesptr, data);
}

uint32_t USBHost::InTransfer(EpInfo *pep, uint32_t nak_limit, uint8_t *nbytesptr, uint8_t* data) {
	uint32_t rcode = 0;
	uint32_t pktsize = 0;

	uint32_t nbytes = *nbytesptr;
	uint32_t maxpktsize = pep->maxPktSize;

	*nbytesptr = 0;
	//set toggle value
	if(pep->bmRcvToggle)
		USB->HOST.HostPipe[pep->epAddr].PSTATUSSET.reg = USB_HOST_PSTATUSSET_DTGL;
	else
		USB->HOST.HostPipe[pep->epAddr].PSTATUSCLR.reg = USB_HOST_PSTATUSCLR_DTGL;

	usb_pipe_table[pep->epAddr].HostDescBank[0].ADDR.reg = (uint32_t)data;

	// use a 'break' to exit this loop
	while (1) {
   		/* get pipe config from setting register */
 		usb_pipe_table[pep->epAddr].HostDescBank[0].ADDR.reg += pktsize;

		rcode = dispatchPkt(tokIN, pep->epAddr, nak_limit); //IN packet to EP-'endpoint'. Function takes care of NAKS.
		if(rcode == USB_ERROR_DATATOGGLE) {
            // yes, we flip it wrong here so that next time it is actually correct!
            //pep->bmRcvToggle = (regRd(rHRSL) & bmSNDTOGRD) ? 0 : 1;
			pep->bmRcvToggle = USB_HOST_DTGL(pep->epAddr);
            //set toggle value
			if(pep->bmRcvToggle)
				USB->HOST.HostPipe[pep->epAddr].PSTATUSSET.reg = USB_HOST_PSTATUSSET_DTGL;
			else
				USB->HOST.HostPipe[pep->epAddr].PSTATUSCLR.reg = USB_HOST_PSTATUSCLR_DTGL;
            continue;
		}
        if(rcode) {
                uhd_freeze_pipe(pep->epAddr);
                //printf(">>>>>>>> Problem! dispatchPkt %2.2x\r\n", rcode);
                return(rcode);// break; //should be 0, indicating ACK. Else return error code.
        }
        /* check for RCVDAVIRQ and generate error if not present */
        /* the only case when absence of RCVDAVIRQ makes sense is when toggle error occurred. Need to add handling for that */
				
		pktsize = uhd_byte_count(pep->epAddr); // Number of received bytes
		
		USB->HOST.HostPipe[pep->epAddr].PSTATUSCLR.reg = USB_HOST_PSTATUSCLR_BK0RDY;
        
		//printf("Got %i bytes \r\n", pktsize);
		// This would be OK, but...
		//assert(pktsize <= nbytes);
		if(pktsize > nbytes) {
			// This can happen. Use of assert on Arduino locks up the Arduino.
			// So I will trim the value, and hope for the best.
			//printf(">>>>>>>> Problem! Wanted %i bytes but got %i.\r\n", nbytes, pktsize);
			pktsize = nbytes;
		}

		int16_t mem_left = (int16_t)nbytes - *((int16_t*)nbytesptr);

		if(mem_left < 0)
			mem_left = 0;

        //data = bytesRd(rRCVFIFO, ((pktsize > mem_left) ? mem_left : pktsize), data);

        //regWr(rHIRQ, bmRCVDAVIRQ); // Clear the IRQ & free the buffer
        *nbytesptr += pktsize;// add this packet's byte count to total transfer length

        /* The transfer is complete under two conditions:           */
        /* 1. The device sent a short packet (L.T. maxPacketSize)   */
        /* 2. 'nbytes' have been transferred.                       */
        if((pktsize < maxpktsize) || (*nbytesptr >= nbytes)) // have we transferred 'nbytes' bytes?
		{					 
            // Save toggle value
            pep->bmRcvToggle = USB_HOST_DTGL(pep->epAddr);
            //printf("\r\n");
            rcode = 0;
            break;
		} // if
	} //while( 1 )
	uhd_freeze_pipe(pep->epAddr);
	return ( rcode);
}

/* OUT transfer to arbitrary endpoint. Handles multiple packets if necessary. Transfers 'nbytes' bytes. */
/* Handles NAK bug per Maxim Application Note 4000 for single buffer transfer   */

/* rcode 0 if no errors. rcode 01-0f is relayed from HRSL                       */
uint32_t USBHost::outTransfer(uint32_t addr, uint32_t ep, uint32_t nbytes, uint8_t* data) {
	EpInfo *pep = NULL;
	uint32_t nak_limit = 0;

	uint32_t rcode = SetPipeAddress(addr, ep, &pep, nak_limit);

	if(rcode)
		return rcode;

	return OutTransfer(pep, nak_limit, nbytes, data);
}

uint32_t USBHost::OutTransfer(EpInfo *pep, uint32_t nak_limit, uint32_t nbytes, uint8_t *data) {
	uint32_t rcode = 0, retry_count;
	uint8_t *data_p = data; //local copy of the data pointer
	uint32_t bytes_tosend, nak_count;
	uint32_t bytes_left = nbytes;
	uint8_t buf[64];
	uint8_t i;

	uint32_t maxpktsize = pep->maxPktSize;

    if(maxpktsize < 1 || maxpktsize > 64)
		return USB_ERROR_INVALID_MAX_PKT_SIZE;

	for( i=0; i<nbytes; i++) {
		buf[i] = data[i];
	}
	//unsigned long timeout = millis() + USB_XFER_TIMEOUT;

	//set toggle value
	if(pep->bmSndToggle)
		USB->HOST.HostPipe[pep->epAddr].PSTATUSSET.reg = USB_HOST_PSTATUSSET_DTGL;
	else
		USB->HOST.HostPipe[pep->epAddr].PSTATUSCLR.reg = USB_HOST_PSTATUSCLR_DTGL;

	while(bytes_left) {
		retry_count = 0;
		nak_count = 0;
		bytes_tosend = (bytes_left >= maxpktsize) ? maxpktsize : bytes_left;
		UHD_Pipe_Write(pep->epAddr, bytes_tosend, buf); //filling output FIFO

		//set number of bytes
		//dispatch packet
		//wait for the completion IRQ
		//clear IRQ

		rcode = dispatchPkt(tokOUT, pep->epAddr, nak_limit);
		if (rcode)
		{
			switch(rcode) {
				case USB_ERRORFLOW:
					nak_count++;
					if(nak_limit && (nak_count == nak_limit))
							goto breakout;
					return ( rcode);
					break;
				case USB_ERRORTIMEOUT:
					retry_count++;
					if(retry_count == USB_RETRY_LIMIT)
							goto breakout;
					return ( rcode);
					break;
				case USB_ERROR_DATATOGGLE:
					// yes, we flip it wrong here so that next time it is actually correct!
					pep->bmSndToggle = USB_HOST_DTGL(pep->epAddr);
					//set toggle value
					if(pep->bmSndToggle)
						USB->HOST.HostPipe[pep->epAddr].PSTATUSSET.reg = USB_HOST_PSTATUSSET_DTGL;
					else
						USB->HOST.HostPipe[pep->epAddr].PSTATUSCLR.reg = USB_HOST_PSTATUSCLR_DTGL;
					break;
				default:
						goto breakout;
			}//switch( rcode
		}

		bytes_left -= bytes_tosend;
		data_p += bytes_tosend;
    }//while( bytes_left...
breakout:

    pep->bmSndToggle = USB_HOST_DTGL(pep->epAddr);
    return ( rcode); //should be 0 in all cases
}

/* dispatch USB packet. Assumes peripheral address is set and relevant buffer is loaded/empty       */
/* If NAK, tries to re-send up to nak_limit times                                                   */
/* If nak_limit == 0, do not count NAKs, exit after timeout                                         */
/* If bus timeout, re-sends up to USB_RETRY_LIMIT times                                             */

/* return codes 0x00-0x0f are HRSLT( 0x00 being success ), 0xff means timeout                       */
uint32_t USBHost::dispatchPkt(uint32_t token, uint32_t epAddr, uint32_t nak_limit) {
	uint32_t timeout = millis() + USB_XFER_TIMEOUT;
	uint32_t nak_count = 0, retry_count=0;
	uint32_t rcode = USB_ERROR_TRANSFER_TIMEOUT;

	TRACE_USBHOST(printf("     => dispatchPkt token=%lu pipe=%lu nak_limit=%lu\r\n", token, epAddr, nak_limit);)

	UHD_Pipe_Send(epAddr, token); //launch the transfer

	// Check timeout but don't hold timeout if VBUS is lost
	while ((timeout > millis()) && (UHD_GetVBUSState() == UHD_STATE_CONNECTED))
	{
		// Wait for transfer completion
		if (UHD_Pipe_Is_Transfer_Complete(epAddr, token))
		{
			return 0;
		}

		//case hrNAK:
		if((USB->HOST.HostPipe[epAddr].PINTFLAG.reg & USB_HOST_PINTFLAG_TRFAIL) ) {
			USB->HOST.HostPipe[epAddr].PINTFLAG.reg = USB_HOST_PINTFLAG_TRFAIL;
			nak_count++;
			if(nak_limit && (nak_count == nak_limit)) {
				rcode = USB_ERRORFLOW;
				return (rcode);
			}
		}

		//case hrNAK:
		if( (usb_pipe_table[epAddr].HostDescBank[0].STATUS_BK.reg & USB_ERRORFLOW ) ) {
			nak_count++;
			if(nak_limit && (nak_count == nak_limit)) {
				rcode = USB_ERRORFLOW;
				return (rcode);
			}
		}

		//case hrTIMEOUT:
		if(usb_pipe_table[epAddr].HostDescBank[0].STATUS_PIPE.reg & USB_ERRORTIMEOUT) {
			retry_count++;
			if(retry_count == USB_RETRY_LIMIT)	{
				rcode = USB_ERRORTIMEOUT;
				return (rcode);
			}
		}

		if( (usb_pipe_table[epAddr].HostDescBank[0].STATUS_PIPE.reg & USB_ERROR_DATATOGGLE ) ) {
			rcode = USB_ERROR_DATATOGGLE;
			return (rcode);
		}
	}

	return rcode;
}

/* USB main task. Performs enumeration/cleanup */
void USBHost::Task(void) //USB state machine
{
	uint32_t rcode = 0;
	volatile uint32_t tmpdata = 0;
	static uint32_t delay = 0;
	//USB_DEVICE_DESCRIPTOR buf;
	uint32_t lowspeed = 0;

	// Update USB task state on Vbus change
	tmpdata = UHD_GetVBUSState();

	/* modify USB task state if Vbus changed */
	switch (tmpdata)
	{
		case UHD_STATE_ERROR: //illegal state
			usb_task_state = USB_DETACHED_SUBSTATE_ILLEGAL;
			lowspeed = 0;
			break;

		case UHD_STATE_DISCONNECTED: // Disconnected state
			if ((usb_task_state & USB_STATE_MASK) != USB_STATE_DETACHED)
				usb_task_state = USB_DETACHED_SUBSTATE_INITIALIZE;
			lowspeed = 0;
			break;

		case UHD_STATE_CONNECTED: // Attached state
			if ((usb_task_state & USB_STATE_MASK) == USB_STATE_DETACHED) {
				delay = millis() + USB_SETTLE_DELAY;
				usb_task_state = USB_ATTACHED_SUBSTATE_SETTLE;
			}
			break;
	}// switch( tmpdata

	// Poll connected devices (if required)
	for (uint32_t i = 0; i < USB_NUMDEVICES; ++i)
		if (devConfig[i])
			rcode = devConfig[i]->Poll();

	// Perform USB enumeration stage and clean up
	switch (usb_task_state) {
		case USB_DETACHED_SUBSTATE_INITIALIZE:
			TRACE_USBHOST(printf(" + USB_DETACHED_SUBSTATE_INITIALIZE\r\n");)

			// Init USB stack and driver
			UHD_Init();

			// Free all USB resources
			for (uint32_t i = 0; i < USB_NUMDEVICES; ++i)
			if (devConfig[i])
				rcode = devConfig[i]->Release();

			usb_task_state = USB_DETACHED_SUBSTATE_WAIT_FOR_DEVICE;
			break;
		case USB_DETACHED_SUBSTATE_WAIT_FOR_DEVICE:  //just sit here
			// Nothing to do
			break;
		case USB_DETACHED_SUBSTATE_ILLEGAL:  //just sit here
			// Nothing to do
			break;
		case USB_ATTACHED_SUBSTATE_SETTLE: // Settle time for just attached device
			if((long)(millis() - delay) >= 0L)
				usb_task_state = USB_ATTACHED_SUBSTATE_RESET_DEVICE;
			break;
		case USB_ATTACHED_SUBSTATE_RESET_DEVICE:
			TRACE_USBHOST(printf(" + USB_ATTACHED_SUBSTATE_RESET_DEVICE\r\n");)
			UHD_BusReset();  //issue bus reset
			usb_task_state = USB_ATTACHED_SUBSTATE_WAIT_RESET_COMPLETE;
			break;
		case USB_ATTACHED_SUBSTATE_WAIT_RESET_COMPLETE:
			if (Is_uhd_reset_sent())
			{
				TRACE_USBHOST(printf(" + USB_ATTACHED_SUBSTATE_WAIT_RESET_COMPLETE\r\n");)

				// Clear Bus Reset flag
				uhd_ack_reset_sent();

				// Enable Start Of Frame generation
				uhd_enable_sof();

				usb_task_state = USB_ATTACHED_SUBSTATE_WAIT_SOF;

				// Wait 20ms after Bus Reset (USB spec)
				delay = millis() + 20;
			}
			break;
		case USB_ATTACHED_SUBSTATE_WAIT_SOF:
			// Wait for SOF received first
			if (Is_uhd_sof())
			{
				if (delay < millis())
				{
					TRACE_USBHOST(printf(" + USB_ATTACHED_SUBSTATE_WAIT_SOF\r\n");)

					// 20ms waiting elapsed
					usb_task_state = USB_STATE_CONFIGURING;
				}
			}
			break;
		case USB_STATE_CONFIGURING:
			TRACE_USBHOST(printf(" + USB_STATE_CONFIGURING\r\n");)
			rcode = Configuring(0, 0, lowspeed);

			if (rcode) {
				TRACE_USBHOST(printf("/!\\ USBHost::Task : USB_STATE_CONFIGURING failed with code: %lu\r\n", rcode);)
				if (rcode != USB_DEV_CONFIG_ERROR_DEVICE_INIT_INCOMPLETE) {
					usb_error = rcode;
					usb_task_state = USB_STATE_ERROR;
				}
			}
			else {
				usb_task_state = USB_STATE_RUNNING;
				TRACE_USBHOST(printf(" + USB_STATE_RUNNING\r\n");)
			}
			break;
		case USB_STATE_RUNNING:
			break;
		case USB_STATE_ERROR:
			break;
	} // switch( usb_task_state )
}

uint32_t USBHost::DefaultAddressing(uint32_t parent, uint32_t port, uint32_t lowspeed) {
        //uint8_t		buf[12];
	uint32_t rcode = 0;
	UsbDeviceDefinition *p0 = NULL, *p = NULL;

	// Get pointer to pseudo device with address 0 assigned
	p0 = addrPool.GetUsbDevicePtr(0);

	if(!p0)
		return USB_ERROR_ADDRESS_NOT_FOUND_IN_POOL;

	if(!p0->epinfo)
		return USB_ERROR_EPINFO_IS_NULL;

	p0->lowspeed = (lowspeed) ? 1 : 0;

	// Allocate new address according to device class
	uint32_t bAddress = addrPool.AllocAddress(parent, 0, port);

	if(!bAddress)
		return USB_ERROR_OUT_OF_ADDRESS_SPACE_IN_POOL;

	p = addrPool.GetUsbDevicePtr(bAddress);

	if(!p)
		return USB_ERROR_ADDRESS_NOT_FOUND_IN_POOL;

	p->lowspeed = lowspeed;

	// Assign new address to the device
	rcode = setAddr(0, 0, bAddress);

	if(rcode) {
		TRACE_USBHOST(printf("/!\\ USBHost::DefaultAddressing : Set address failed with code: %lu\r\n", rcode);)
		addrPool.FreeAddress(bAddress);
		bAddress = 0;
		return rcode;
	}
	return 0;
}

uint32_t USBHost::AttemptConfig(uint32_t driver, uint32_t parent, uint32_t port, uint32_t lowspeed) {
        //printf("AttemptConfig: parent = %i, port = %i\r\n", parent, port);
        uint8_t retries = 0;

again:
        uint8_t rcode = devConfig[driver]->ConfigureDevice(parent, port, lowspeed);
        if(rcode == USB_ERROR_CONFIG_REQUIRES_ADDITIONAL_RESET) {
                if(parent == 0) {
                        // Send a bus reset on the root interface.
                        //regWr(rHCTL, bmBUSRST); //issue bus reset
                        UHD_BusReset();
                        delay(102); // delay 102ms, compensate for clock inaccuracy.
                } else {
                        // reset parent port
                        devConfig[parent]->ResetHubPort(port);
                }
        } else if(rcode != 0x00/*hrJERR*/ && retries < 3) { // Some devices returns this when plugged in - trying to initialize the device again usually works
                delay(100);
                retries++;
                goto again;
        } else if(rcode)
                return rcode;

        rcode = devConfig[driver]->Init(parent, port, lowspeed);
        if(rcode != 0x00/*hrJERR*/ && retries < 3) { // Some devices returns this when plugged in - trying to initialize the device again usually works
                delay(100);
                retries++;
                goto again;
        }
        if(rcode) {
                // Issue a bus reset, because the device may be in a limbo state
                if(parent == 0) {
                        // Send a bus reset on the root interface.
                        //regWr(rHCTL, bmBUSRST); //issue bus reset
                        UHD_BusReset();
                        delay(102); // delay 102ms, compensate for clock inaccuracy.
                } else {
                        // reset parent port
                        devConfig[parent]->ResetHubPort(port);
                }
        }
        return rcode;
}

/*
 * This is broken. We need to enumerate differently.
 * It causes major problems with several devices if detected in an unexpected order.
 *
 *
 * Oleg - I wouldn't do anything before the newly connected device is considered sane.
 * i.e.(delays are not indicated for brevity):
 * 1. reset
 * 2. GetDevDescr();
 * 3a. If ACK, continue with allocating address, addressing, etc.
 * 3b. Else reset again, count resets, stop at some number (5?).
 * 4. When max.number of resets is reached, toggle power/fail
 * If desired, this could be modified by performing two resets with GetDevDescr() in the middle - however, from my experience, if a device answers to GDD()
 * it doesn't need to be reset again
 * New steps proposal:
 * 1: get address pool instance. exit on fail
 * 2: pUsb->getDevDescr(0, 0, constBufSize, (uint8_t*)buf). exit on fail.
 * 3: bus reset, 100ms delay
 * 4: set address
 * 5: pUsb->setEpInfoEntry(bAddress, 1, epInfo), exit on fail
 * 6: while (configurations) {
 *              for(each configuration) {
 *                      for (each driver) {
 *                              6a: Ask device if it likes configuration. Returns 0 on OK.
 *                                      If successful, the driver configured device.
 *                                      The driver now owns the endpoints, and takes over managing them.
 *                                      The following will need codes:
 *                                          Everything went well, instance consumed, exit with success.
 *                                          Instance already in use, ignore it, try next driver.
 *                                          Not a supported device, ignore it, try next driver.
 *                                          Not a supported configuration for this device, ignore it, try next driver.
 *                                          Could not configure device, fatal, exit with fail.
 *                      }
 *              }
 *    }
 * 7: for(each driver) {
 *      7a: Ask device if it knows this VID/PID. Acts exactly like 6a, but using VID/PID
 * 8: if we get here, no driver likes the device plugged in, so exit failure.
 *
 */
uint32_t USBHost::Configuring(uint32_t parent, uint32_t port, uint32_t lowspeed) {
       //uint32_t bAddress = 0;
        //printf("Configuring: parent = %i, port = %i\r\n", parent, port);
        uint32_t devConfigIndex;
        uint32_t rcode = 0;
        uint8_t buf[sizeof (USB_DEVICE_DESCRIPTOR)];
        USB_DEVICE_DESCRIPTOR *udd = reinterpret_cast<USB_DEVICE_DESCRIPTOR *>(buf);
        UsbDeviceDefinition *p = NULL;
        EpInfo *oldep_ptr = NULL;
        EpInfo epInfo;

        epInfo.epAddr = 0;
        epInfo.maxPktSize = 8;
        epInfo.bmSndToggle = 0;
        epInfo.bmRcvToggle = 0;
        epInfo.bmNakPower = USB_NAK_MAX_POWER;

        //delay(2000);
        AddressPool &addrPool = GetAddressPool();
        // Get pointer to pseudo device with address 0 assigned
        p = addrPool.GetUsbDevicePtr(0);
        if(!p) {
                //printf("Configuring error: USB_ERROR_ADDRESS_NOT_FOUND_IN_POOL\r\n");
                return USB_ERROR_ADDRESS_NOT_FOUND_IN_POOL;
        }

        // Save old pointer to EP_RECORD of address 0
        oldep_ptr = p->epinfo;

        // Temporary assign new pointer to epInfo to p->epinfo in order to
        // avoid toggle inconsistence

        p->epinfo = &epInfo;

        p->lowspeed = lowspeed;
        // Get device descriptor
        rcode = getDevDescr(0, 0, sizeof (USB_DEVICE_DESCRIPTOR), (uint8_t*)buf);
        // The first GetDescriptor give us the endpoint 0 max packet size.
        epInfo.maxPktSize = buf[7];
        // Restore p->epinfo
        p->epinfo = oldep_ptr;

        if(rcode) {
                //printf("Configuring error: Can't get USB_DEVICE_DESCRIPTOR\r\n");
                return rcode;
        }

        // to-do?
        // Allocate new address according to device class
        //bAddress = addrPool.AllocAddress(parent, false, port);

        //if (!bAddress)
        //        return USB_ERROR_OUT_OF_ADDRESS_SPACE_IN_POOL;
        uint16_t vid = udd->idVendor;
        uint16_t pid = udd->idProduct;
        uint8_t klass = udd->bDeviceClass;

        // Attempt to configure if VID/PID or device class matches with a driver
        for(devConfigIndex = 0; devConfigIndex < USB_NUMDEVICES; devConfigIndex++) {
                if(!devConfig[devConfigIndex]) continue; // no driver
                if(devConfig[devConfigIndex]->GetAddress()) continue; // consumed
                if(devConfig[devConfigIndex]->VIDPIDOK(vid, pid) || devConfig[devConfigIndex]->DEVCLASSOK(klass)) {
                        rcode = AttemptConfig(devConfigIndex, parent, port, lowspeed);
                        if(rcode != USB_DEV_CONFIG_ERROR_DEVICE_NOT_SUPPORTED)
                                break;
                }
        }

        if(devConfigIndex < USB_NUMDEVICES) {
                return rcode;
        }


        // blindly attempt to configure
        for(devConfigIndex = 0; devConfigIndex < USB_NUMDEVICES; devConfigIndex++) {
                if(!devConfig[devConfigIndex]) continue;
                if(devConfig[devConfigIndex]->GetAddress()) continue; // consumed
                if(devConfig[devConfigIndex]->VIDPIDOK(vid, pid) || devConfig[devConfigIndex]->DEVCLASSOK(klass)) continue; // If this is true it means it must have returned USB_DEV_CONFIG_ERROR_DEVICE_NOT_SUPPORTED above
                rcode = AttemptConfig(devConfigIndex, parent, port, lowspeed);

                //printf("ERROR ENUMERATING %2.2x\r\n", rcode);
                if(!(rcode == USB_DEV_CONFIG_ERROR_DEVICE_NOT_SUPPORTED || rcode == USB_ERROR_CLASS_INSTANCE_ALREADY_IN_USE)) {
                        // in case of an error dev_index should be reset to 0
                        //		in order to start from the very beginning the
                        //		next time the program gets here
                        //if (rcode != USB_DEV_CONFIG_ERROR_DEVICE_INIT_INCOMPLETE)
                        //        devConfigIndex = 0;
                        return rcode;
                }
	}
        // if we get here that means that the device class is not supported by any of registered classes
	rcode = DefaultAddressing(parent, port, lowspeed);

	return rcode;
}

uint32_t USBHost::ReleaseDevice(uint32_t addr) {
	if(!addr)
		return 0;

	for(uint32_t i = 0; i < USB_NUMDEVICES; i++) {
                if(!devConfig[i]) continue;
		if(devConfig[i]->GetAddress() == addr)
			return devConfig[i]->Release();
	}
	return 0;
}

//get device descriptor

uint32_t USBHost::getDevDescr(uint32_t addr, uint32_t ep, uint32_t nbytes, uint8_t* dataptr) {
    return (ctrlReq(addr, ep, bmREQ_GET_DESCR, USB_REQUEST_GET_DESCRIPTOR, 0x00, USB_DESCRIPTOR_DEVICE, 0x0000, nbytes, nbytes, dataptr, 0));
}
//get configuration descriptor

uint32_t USBHost::getConfDescr(uint32_t addr, uint32_t ep, uint32_t nbytes, uint32_t conf, uint8_t* dataptr) {
	return (ctrlReq(addr, ep, bmREQ_GET_DESCR, USB_REQUEST_GET_DESCRIPTOR, conf, USB_DESCRIPTOR_CONFIGURATION, 0x0000, nbytes, nbytes, dataptr, 0));
}

/* Requests Configuration Descriptor. Sends two Get Conf Descr requests. The first one gets the total length of all descriptors, then the second one requests this
 total length. The length of the first request can be shorter ( 4 bytes ), however, there are devices which won't work unless this length is set to 9 */
uint32_t USBHost::getConfDescr(uint32_t addr, uint32_t ep, uint32_t conf, USBReadParser *p) {
	const uint32_t bufSize = 64;
	uint8_t buf[bufSize];
	USB_CONFIGURATION_DESCRIPTOR *ucd = reinterpret_cast<USB_CONFIGURATION_DESCRIPTOR *>(buf);

	uint32_t ret = getConfDescr(addr, ep, 9, conf, buf);

	if(ret)
		return ret;

        uint32_t total = ucd->wTotalLength;

        //USBTRACE2("\r\ntotal conf.size:", total);

        return (ctrlReq(addr, ep, bmREQ_GET_DESCR, USB_REQUEST_GET_DESCRIPTOR, conf, USB_DESCRIPTOR_CONFIGURATION, 0x0000, total, bufSize, buf, p));
}

//get string descriptor

uint32_t USBHost::getStrDescr(uint32_t addr, uint32_t ep, uint32_t nbytes, uint32_t index, uint32_t langid, uint8_t* dataptr) {
    return (ctrlReq(addr, ep, bmREQ_GET_DESCR, USB_REQUEST_GET_DESCRIPTOR, index, USB_DESCRIPTOR_STRING, langid, nbytes, nbytes, dataptr, 0));
}
//set address

uint32_t USBHost::setAddr(uint32_t oldaddr, uint32_t ep, uint32_t newaddr) {
        uint32_t rcode = ctrlReq(oldaddr, ep, bmREQ_SET, USB_REQUEST_SET_ADDRESS, newaddr, 0x00, 0x0000, 0x0000, 0x0000, NULL, NULL);
        //delay(2); //per USB 2.0 sect.9.2.6.3
        delay(300); // Older spec says you should wait at least 200ms
        return rcode;
        //return ( ctrlReq(oldaddr, ep, bmREQ_SET, USB_REQUEST_SET_ADDRESS, newaddr, 0x00, 0x0000, 0x0000, 0x0000, NULL, NULL));
}
//set configuration

uint32_t USBHost::setConf(uint32_t addr, uint32_t ep, uint32_t conf_value) {
        return ( ctrlReq(addr, ep, bmREQ_SET, USB_REQUEST_SET_CONFIGURATION, conf_value, 0x00, 0x0000, 0x0000, 0x0000, NULL, NULL));
}

//#endif //ARDUINO_SAMD_ZERO
