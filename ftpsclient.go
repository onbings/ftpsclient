// Copyright 2014 OnBings. All rights reserved.
// Use of this source code is governed by a APACHE-style
// license that can be found in the LICENSE file.

/*
	Package FtpsClient implements a basic ftp(s) client which can be used to connect
	an application to a ftp server. Only a small subset of the full FTP/FTPS specification
	is supported

	It is based on the word made by:
	- jlaffaye 		github.com/jlaffaye/ftp
	- smallfish 		github.com/smallfish/ftp
	- Marco Beierer 	github.com/webguerilla/ftps

	And has some other feature such as:
	- Refactored with onbings coding covention
	- Add secure/unsecure mode
	- Add timeout support
	- Add generic Ftp control send command function (SendFtpCommand) to be able to send SITE, NOOP,... ftp command)

	Usage

*/
package ftpsclient

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/textproto"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidParameter = errors.New("Ftps: Invalid parameter")
	ErrNotConnected     = errors.New("Ftps: Connection is not established")
	ErrPasv             = errors.New("Ftps: Invalid PASV response format")
	ErrIoError          = errors.New("Ftps: File transfer not complete")
	ErrLineFormat       = errors.New("Ftps: Unsupported line format")
	ErrDirEntry         = errors.New("Ftps: Unknown directory entry type")
	ErrInvalidLogin     = errors.New("Ftps: Invalid loging")
	ErrInvalidDirectory = errors.New("Ftps: Invalid directory")
	ErrNotDisconnected  = errors.New("Ftps: Can't disconnect")
	ErrSecure           = errors.New("Ftps: Secure protocol error")
)

type DIRENTRYTYPE int

const (
	DIRENTRYTYPE_FILE DIRENTRYTYPE = iota
	DIRENTRYTYPE_FOLDER
	DIRENTRYTYPE_LINK
)

type DirEntry struct {
	Type_E   DIRENTRYTYPE
	Name_S   string
	Ext_S    string
	Size_U64 uint64
	Time_X   time.Time
}
type FtpsClientParam struct {
	//public
	Id_U32                  uint32
	LoginName_S             string
	LoginPassword_S         string
	InitialDirectory_S      string
	SecureFtp_B             bool
	TargetHost_S            string
	TargetPort_U16          uint16
	Debug_B                 bool
	TlsConfig_X             tls.Config
	ConnectTimeout_S64      time.Duration
	CtrlTimeout_S64         time.Duration
	DataTimeout_S64         time.Duration
	CtrlReadBufferSize_U32  uint32
	CtrlWriteBufferSize_U32 uint32
	DataReadBufferSize_U32  uint32
	DataWriteBufferSize_U32 uint32
}

type FtpsClient struct {
	FtpsParam_X FtpsClientParam

	ctrlConnection_I  net.Conn
	dataConnection_I  net.Conn
	textProtocolPtr_X *textproto.Conn
}

type ConBufferSetter interface {
	SetReadBuffer(bytes int) error
	SetWriteBuffer(bytes int) error
}

func setConBufferSize(_Connection_I net.Conn, _ReadBufferSize_U32 uint32, _WriteBufferSize_U32 uint32) (rRts error) {
	rRts = nil
	Obj_I, Ok_B := _Connection_I.(ConBufferSetter)
	if Ok_B {
		if _ReadBufferSize_U32 != 0 {
			Obj_I.SetReadBuffer(int(_ReadBufferSize_U32))
		}
		if _WriteBufferSize_U32 != 0 {
			Obj_I.SetWriteBuffer(int(_WriteBufferSize_U32))
		}
	}
	return
}
func NewFtpsClient(_FtpsClientParamPtr_X *FtpsClientParam) *FtpsClient {
	p := new(FtpsClient)
	p.FtpsParam_X = *_FtpsClientParamPtr_X
	return p
}

func (this *FtpsClient) Connect() (rRts error) {
	var Sts error

	rRts = ErrNotConnected
	this.ctrlConnection_I, Sts = net.DialTimeout("tcp4", fmt.Sprintf("%s:%d", this.FtpsParam_X.TargetHost_S, this.FtpsParam_X.TargetPort_U16), this.FtpsParam_X.ConnectTimeout_S64)
	if Sts == nil {
		Sts = setConBufferSize(this.ctrlConnection_I, this.FtpsParam_X.CtrlReadBufferSize_U32, this.FtpsParam_X.CtrlWriteBufferSize_U32)
		if Sts == nil {
			this.textProtocolPtr_X = textproto.NewConn(this.ctrlConnection_I)
			_, _, Sts = this.readFtpServerResponse(220)

			if Sts == nil {
				if this.FtpsParam_X.SecureFtp_B {
					rRts = ErrSecure
					_, _, Sts = this.sendRequestToFtpServer("AUTH TLS", 234)
					if Sts == nil {
						this.ctrlConnection_I = this.upgradeConnectionToTLS(this.ctrlConnection_I)
						this.textProtocolPtr_X = textproto.NewConn(this.ctrlConnection_I)
					}
				}
			}

			if Sts == nil {
				rRts = ErrInvalidLogin
				_, _, Sts = this.sendRequestToFtpServer(fmt.Sprintf("USER %s", this.FtpsParam_X.LoginName_S), 331)
				if Sts == nil {
					_, _, Sts = this.sendRequestToFtpServer(fmt.Sprintf("PASS %s", this.FtpsParam_X.LoginPassword_S), 230)
					if Sts == nil {
						rRts = ErrInvalidParameter
						_, _, Sts = this.sendRequestToFtpServer("TYPE I", 200)
						if Sts == nil {
							rRts = ErrInvalidDirectory
							_, _, Sts = this.sendRequestToFtpServer(fmt.Sprintf("CWD %s", this.FtpsParam_X.InitialDirectory_S), 250)

							if Sts == nil {
								if this.FtpsParam_X.SecureFtp_B {
									rRts = ErrSecure
									_, _, Sts = this.sendRequestToFtpServer("PBSZ 0", 200)
									if Sts == nil {
										_, _, Sts = this.sendRequestToFtpServer("PROT P", 200) // encrypt data connection
									}
								}
							}
						}
					}
				}
			}
			if Sts == nil {
				rRts = nil
			}
		}
	}
	return
}

func (this *FtpsClient) GetWorkingDirectory() (rDirectory_S string, rRts error) {

	_, rDirectory_S, rRts = this.sendRequestToFtpServer("PWD", 257)
	if rRts == nil {
		StartPos_i := strings.Index(rDirectory_S, "\"")
		EndPos_i := strings.LastIndex(rDirectory_S, "\"")
		//		fmt.Printf("GetWorkingDirectory '%s' %d %d\n", rDirectory_S, StartPos_i, EndPos_i)
		if StartPos_i == -1 || EndPos_i == -1 || StartPos_i >= EndPos_i {
			rRts = ErrLineFormat
		} else {
			rDirectory_S = rDirectory_S[StartPos_i+1 : EndPos_i]
		}
		//		fmt.Printf("GetWorkingDirectory->'%s'\n", rDirectory_S)

	}

	return
}

func (this *FtpsClient) ChangeWorkingDirectory(_Path_S string) (rRts error) {

	_, _, rRts = this.sendRequestToFtpServer(fmt.Sprintf("CWD %s", _Path_S), 250)
	return
}

func (this *FtpsClient) MakeDirectory(_Path_S string) (rRts error) {

	_, _, rRts = this.sendRequestToFtpServer(fmt.Sprintf("MKD %s", _Path_S), 257)
	return
}

func (this *FtpsClient) DeleteFile(_Path_S string) (rRts error) {

	_, _, rRts = this.sendRequestToFtpServer(fmt.Sprintf("DELE %s", _Path_S), 250)
	return
}

func (this *FtpsClient) RemoveDirectory(_Path_S string) (rRts error) {

	_, _, rRts = this.sendRequestToFtpServer(fmt.Sprintf("RMD %s", _Path_S), 250)
	return
}

func (this *FtpsClient) SendFtpCtrlCommand(_FtpCommand_S string, _ExpectedReplyCode_i int) (rReplyCode_i int, rReplyMessage_S string, rRts error) {
	rReplyCode_i, rReplyMessage_S, rRts = this.sendRequestToFtpServer(_FtpCommand_S, _ExpectedReplyCode_i)
	return
}

func (this *FtpsClient) OpenFtpDataChannel(_FtpCommand_S string, _ExpectedReplyCode_i int) (rReplyCode_i int, rReplyMessage_S string, rRts error) {
	rRts = this.sendRequestToFtpServerDataConn(_FtpCommand_S, _ExpectedReplyCode_i)
	return
}
func (this *FtpsClient) ReadFtpDataChannel(_DataArray_U8 []uint8) (rWaitDuration_S64 time.Duration, rIoDuration_S64 time.Duration, rNbRead_i int, rRts error) {
	var NbRead_i int
	var StartWaitTime_X, StartIoTime_X time.Time
	var FirstIo_B bool

	StartWaitTime_X = time.Now()
	NbMaxToRead_i := len(_DataArray_U8)
	rNbRead_i = 0
	rRts = this.dataConnection_I.SetDeadline(time.Now().Add(this.FtpsParam_X.DataTimeout_S64))
	//	fmt.Printf("now %v to %v\n", time.Now(), time.Now().Add(this.FtpsParam_X.DataTimeout_S64))

	if rRts == nil {
		FirstIo_B = true
		for {
			NbRead_i, rRts = this.dataConnection_I.Read(_DataArray_U8[rNbRead_i:])
			if rRts == nil {
				if FirstIo_B {
					StartIoTime_X = time.Now()
					rWaitDuration_S64 = StartIoTime_X.Sub(StartWaitTime_X)
					FirstIo_B = false
				}
				rNbRead_i = rNbRead_i + NbRead_i
				if rNbRead_i >= NbMaxToRead_i {
					//			fmt.Printf("FINAL GOT %d/%d -> cont\n", rNbRead_i, NbMaxToRead_i)
					rIoDuration_S64 = time.Now().Sub(StartIoTime_X)
					break
				} else {
					//					fmt.Printf(">>Partial got %d/%d -> cont\n", rNbRead_i, NbMaxToRead_i)
				}

			} else {

				if rRts == io.EOF {
					//					time.Sleep(time.Millisecond * 100)
					//					fmt.Printf("EOF->cont\n")
				} else {
					//				fmt.Printf("%v %d/%d err1 %s\n", time.Now(), rNbRead_i, NbMaxToRead_i, rRts.Error())
					break
				}
			}
		}

	} else {
		fmt.Printf("err2 %s\n", rRts.Error())
	}
	return
}
func (this *FtpsClient) CloseFtpDataChannel() (rReplyCode_i int, rReplyMessage_S string, rRts error) {
	rReplyMessage_S = ""
	rReplyCode_i = 0
	rRts = this.dataConnection_I.Close()
	if rRts == nil {
		rReplyCode_i, rReplyMessage_S, rRts = this.readFtpServerResponse(226)
	}
	return
}
func (this *FtpsClient) List() (rDirEntryArray_X []DirEntry, rRts error) {
	var DirEntryPtr_X *DirEntry
	var Line_S string
	rDirEntryArray_X = nil
	rRts = this.sendRequestToFtpServerDataConn("LIST -a", 150)
	if rRts == nil {
		pReader_O := bufio.NewReader(this.dataConnection_I)
		for {
			Line_S, rRts = pReader_O.ReadString('\n')
			if rRts == nil {
				//if rRts == io.EOF {				break			}

				DirEntryPtr_X, rRts = this.parseEntryLine(Line_S)
				rDirEntryArray_X = append(rDirEntryArray_X, *DirEntryPtr_X)
			} else {
				break
			}

		}
		_, _, rRts = this.readFtpServerResponse(226)
		this.dataConnection_I.Close()
	}

	return
}

func (this *FtpsClient) StoreFile(_RemoteFilepath_S string, _DataArray_U8 []byte) (rRts error) {
	var Count_i int

	rRts = this.sendRequestToFtpServerDataConn(fmt.Sprintf("STOR %s", _RemoteFilepath_S), 150)
	if rRts == nil {
		Count_i, rRts = this.dataConnection_I.Write(_DataArray_U8)
		if rRts == nil {
			if len(_DataArray_U8) != Count_i {
				rRts = ErrIoError
			} else {
				_, _, rRts = this.readFtpServerResponse(226)
			}
		}
		this.dataConnection_I.Close()
	}
	return
}

func (this *FtpsClient) RetrieveFile(_RemoteFilepath_S, _LocalFilepath_S string) (rRts error) {
	var pFile_X *os.File

	rRts = this.sendRequestToFtpServerDataConn(fmt.Sprintf("RETR %s", _RemoteFilepath_S), 150)
	if rRts == nil {
		pFile_X, rRts = os.Create(_LocalFilepath_S)
		if rRts == nil {
			_, rRts = io.Copy(pFile_X, this.dataConnection_I)
			if rRts == nil {
				_, _, rRts = this.readFtpServerResponse(226)
			}
			pFile_X.Close()
		}
		this.dataConnection_I.Close()
	}
	return
}

func (this *FtpsClient) RetrieveData(_RemoteFilepath_S, _LocalFilepath_S string) (rRts error) {
	var pFile_X *os.File

	rRts = this.sendRequestToFtpServerDataConn(fmt.Sprintf("RETR %s", _RemoteFilepath_S), 150)
	if rRts == nil {
		pFile_X, rRts = os.Create(_LocalFilepath_S)
		if rRts == nil {
			_, rRts = io.Copy(pFile_X, this.dataConnection_I)
			if rRts == nil {
				_, _, rRts = this.readFtpServerResponse(226)
			}
			pFile_X.Close()
			this.dataConnection_I.Close()
		}
	}
	return
}

func (this *FtpsClient) Disconnect() (rRts error) {
	_, _, rRts = this.sendRequestToFtpServer("QUIT", 221)
	if rRts == nil {
		rRts = this.ctrlConnection_I.Close()
	}
	return
}
func (this *FtpsClient) isConnEstablished() (rRts error) {
	rRts = ErrNotConnected
	if this.ctrlConnection_I == nil {
		//		panic(rRts.Error())
	} else {
		rRts = nil
	}
	return
}

func (this *FtpsClient) openDataConn(_Port_i int) (rRts error) {
	var Sts error

	rRts = ErrNotConnected
	this.dataConnection_I, Sts = net.DialTimeout("tcp4", fmt.Sprintf("%s:%d", this.FtpsParam_X.TargetHost_S, _Port_i), this.FtpsParam_X.ConnectTimeout_S64)
	if Sts == nil {
		rRts = setConBufferSize(this.dataConnection_I, this.FtpsParam_X.DataReadBufferSize_U32, this.FtpsParam_X.DataWriteBufferSize_U32)
	}
	return
}
func (this *FtpsClient) sendRequestToFtpServer(_Request_S string, _ExpectedReplyCode_i int) (rReplyCode_i int, rReplyMessage_S string, rRts error) {
	rReplyCode_i = 0
	rReplyMessage_S = ""
	rRts = this.isConnEstablished()
	if rRts == nil {
		this.debugInfo("[FTP CMD] " + _Request_S)
		rRts = this.ctrlConnection_I.SetDeadline(time.Now().Add(this.FtpsParam_X.CtrlTimeout_S64))
		if rRts == nil {
			_, rRts = this.textProtocolPtr_X.Cmd(_Request_S)
			if rRts == nil {
				rReplyCode_i, rReplyMessage_S, rRts = this.readFtpServerResponse(_ExpectedReplyCode_i)
			}
		}
	}
	return
}
func (this *FtpsClient) readFtpServerResponse(_ExpectedReplyCode_i int) (rReplyCode_i int, rResponse_S string, rRts error) {
	rReplyCode_i = 0
	rResponse_S = ""
	rRts = this.isConnEstablished()
	if rRts == nil {
		rRts = this.ctrlConnection_I.SetDeadline(time.Now().Add(this.FtpsParam_X.CtrlTimeout_S64))
		if rRts == nil {
			rReplyCode_i, rResponse_S, rRts = this.textProtocolPtr_X.ReadResponse(_ExpectedReplyCode_i)

			this.debugInfo(fmt.Sprintf("[FTP REP] %d/%d (%s)", rReplyCode_i, _ExpectedReplyCode_i, rResponse_S))
		}
	}
	return
}
func (this *FtpsClient) preparePasvConnection() (rPort_i int, rRts error) {
	var ReplyMessage_S string
	var PortPart1_i, PortPart2_i int

	rPort_i = 0
	_, ReplyMessage_S, rRts = this.sendRequestToFtpServer("PASV", 227)
	if rRts == nil {
		StartPos_i := strings.Index(ReplyMessage_S, "(")
		EndPos_i := strings.LastIndex(ReplyMessage_S, ")")

		if StartPos_i == -1 || EndPos_i == -1 {
			rRts = ErrPasv
		} else {
			pPasvData_S := strings.Split(ReplyMessage_S[StartPos_i+1:EndPos_i], ",")

			PortPart1_i, rRts = strconv.Atoi(pPasvData_S[4])
			if rRts == nil {
				PortPart2_i, rRts = strconv.Atoi(pPasvData_S[5])
				if rRts == nil {
					// Recompose port
					rPort_i = int(PortPart1_i)*256 + int(PortPart2_i)
				}
			}
		}
	}
	return
}

func (this *FtpsClient) sendRequestToFtpServerDataConn(_Request_S string, _ExpectedReplyCode_i int) (rRts error) {
	var Port_i int

	Port_i, rRts = this.preparePasvConnection()
	if rRts == nil {
		rRts = this.openDataConn(Port_i)
		if rRts == nil {
			_, _, rRts = this.sendRequestToFtpServer(_Request_S, _ExpectedReplyCode_i)
			if rRts != nil {
				this.dataConnection_I.Close()
				this.dataConnection_I = nil
			} else {
				if this.FtpsParam_X.SecureFtp_B {

					this.dataConnection_I = this.upgradeConnectionToTLS(this.dataConnection_I)
				}
			}

		}
	}

	return
}

func (pFtpsClient_X *FtpsClient) upgradeConnectionToTLS(_Connection_I net.Conn) (rUpgradedConnection net.Conn) {

	var TlsConnectionPtr_X *tls.Conn
	TlsConnectionPtr_X = tls.Client(_Connection_I, &pFtpsClient_X.FtpsParam_X.TlsConfig_X)

	TlsConnectionPtr_X.Handshake()
	rUpgradedConnection = net.Conn(TlsConnectionPtr_X)

	// TODO verify that TLS connection is established

	return
}

func (this *FtpsClient) parseEntryLine(_Line_S string) (rDirEntryPtr_X *DirEntry, rRts error) {
	var Time_S string
	var Size_U64 uint64
	var Time_X time.Time

	//filename can contains space:  _Line_S="-rwx------ 1 user group 16835936256 May 26 06:40 000000B_!RTN3V}H_Train000002             .TRN"
	rDirEntryPtr_X = nil
	//	Field_S := strings.Fields(_Line_S)
	FieldArray_S := strings.SplitN(_Line_S, " ", 9)
	rRts = ErrLineFormat
	if len(FieldArray_S) == 9 {
		FnStartPos_i := strings.LastIndex(_Line_S, FieldArray_S[7])
		if FnStartPos_i >= 0 {
			FnStartPos_i = FnStartPos_i + len(FieldArray_S[7]) - 1
			rDirEntryPtr_X = &DirEntry{}
			/*
				this.debugInfo(fmt.Sprintf("[FTP DBG] line '%s'", _Line_S))
				for i := 0; i < len(FieldArray_S); i++ {
					this.debugInfo(fmt.Sprintf("[FTP DBG] %d: '%s'", i, FieldArray_S[i]))
				}
			*/
			// parse type
			switch FieldArray_S[0][0] {
			case '-':
				rDirEntryPtr_X.Type_E = DIRENTRYTYPE_FILE
			case 'd':
				rDirEntryPtr_X.Type_E = DIRENTRYTYPE_FOLDER
			case 'l':
				rDirEntryPtr_X.Type_E = DIRENTRYTYPE_LINK
			default:
				rRts = ErrDirEntry
			}

			// parse size
			Size_U64, rRts = strconv.ParseUint(FieldArray_S[4], 10, 64)
			if rRts != nil {
				//			this.debugInfo(fmt.Sprintf("[FTP DBG] err '%s'", rRts.Error()))
				rDirEntryPtr_X = nil
			} else {

				rDirEntryPtr_X.Size_U64 = Size_U64
				// parse time
				if strings.Contains(FieldArray_S[7], ":") { // this year
					Year_i, _, _ := time.Now().Date()
					Time_S = fmt.Sprintf("%s %s %s %s GMT", FieldArray_S[6], FieldArray_S[5], strconv.Itoa(Year_i)[2:4], FieldArray_S[7])
				} else { // not this year
					Time_S = fmt.Sprintf("%s %s %s 00:00 GMT", FieldArray_S[6], FieldArray_S[5], FieldArray_S[7][2:4])
				}
				Time_X, rRts = time.Parse("_2 Jan 06 15:04 MST", Time_S)
				if rRts != nil {
					rDirEntryPtr_X = nil
				} else {
					rDirEntryPtr_X.Time_X = Time_X // TODO set timezone

					// parse name
					rDirEntryPtr_X.Name_S = strings.TrimRight(FieldArray_S[8], "\r\n")
					SepIndex_i := strings.LastIndex(rDirEntryPtr_X.Name_S, ".")
					if SepIndex_i >= 0 {
						rDirEntryPtr_X.Ext_S = rDirEntryPtr_X.Name_S[SepIndex_i+1:]
						rDirEntryPtr_X.Name_S = rDirEntryPtr_X.Name_S[:SepIndex_i]
					}
				}
			}
		}
	}
	return
}

func (this *FtpsClient) debugInfo(_Message_S string) {

	if this.FtpsParam_X.Debug_B {
		log.Println(_Message_S)
	}
}
