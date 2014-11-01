FTPS client for Go(lang)
========================
Another Ftp(s) client package for the go language.
This is a non working copy of the package.
It is under finalization with unit test support.

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
	
INSTALL 
========
go get github.com/onbings/ftpsclient

EXAMPLE 
========
```go
package main

import (          
	"fmt"
	"log"
	"github.com/onbings/ftpsclient"
)

func main() {

	var FtpsClientParam_X ftpsclient.FtpsClientParam

	FtpsClientParam_X.Id_U32 = 123
	FtpsClientParam_X.LoginName_S = "mc"
	FtpsClientParam_X.LoginPassword_S = "a"
	FtpsClientParam_X.InitialDirectory_S = "/Seq"
	FtpsClientParam_X.SecureFtp_B = false
	FtpsClientParam_X.TargetHost_S = "127.0.0.1"
	FtpsClientParam_X.TargetPort_U16 = 21
	FtpsClientParam_X.Debug_B = false
	FtpsClientParam_X.TlsConfig_X.InsecureSkipVerify = true
	FtpsClientParam_X.ConnectTimeout_S64 = 2000
	FtpsClientParam_X.CtrlTimeout_S64 = 1000
	FtpsClientParam_X.DataTimeout_S64 = 5000
	FtpsClientParam_X.CtrlReadBufferSize_U32 = 0
	FtpsClientParam_X.CtrlWriteBufferSize_U32 = 0
	FtpsClientParam_X.DataReadBufferSize_U32 = 0x100000
	FtpsClientParam_X.DataWriteBufferSize_U32 = 0x100000

	FtpsClientPtr_X := ftpsclient.NewFtpsClient(&FtpsClientParam_X)
	if FtpsClientPtr_X != nil {
		Err := FtpsClientPtr_X.Connect()
		if Err == nil {
			DirEntryArray_X, Err := FtpsClientPtr_X.List()
			if Err == nil {
				for _, DirEntry_X := range DirEntryArray_X {
					fmt.Printf("(%d): %s.%s %d bytes %s\n", DirEntry_X.Type_E, DirEntry_X.Name_S, DirEntry_X.Ext_S, DirEntry_X.Size_U64, DirEntry_X.Time_X)
				}
				Err = FtpsClientPtr_X.Disconnect()
			}
		}
	}
}
```
