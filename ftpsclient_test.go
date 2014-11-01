// Copyright 2014 OnBings. All rights reserved.
// Use of this source code is governed by a APACHE-style
// license that can be found in the LICENSE file.

/*
	This module implements the 'ftpsclient' package unit test.
	To run theses test please install Filezilla Ftp server (https://filezilla-project.org/download.php?type=server)
	and create a ftp server config with a user name 'mc' with pasword 'a'. The ftp server should exposes a 'Seq'
	directory.

	These unit tests use the 'gocheck' golang test package (https://labix.org/gocheck)

*/
package ftpsclient

import (
	//"fmt"
	. "gopkg.in/check.v1"
	"io/ioutil"
	//"time"
	//"log"
	"testing"
)

//-- GoCheck specific initialization --------------------------------------

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type FtpClientTestSuite struct{}

var _ = Suite(&FtpClientTestSuite{})

//-- GoCheck specific initialization --------------------------------------
const (
	HOST_PATH       = "copyofftpsclient.go.bha"
	LOCAL_PATH      = "ftpsclient.go"
	LOCAL_PATH_RETR = "ftpsclientfromftp.bha"
)

var (
	GL_FtpsClientPtr_X *FtpsClient
)

//Fixtures are available by using one or more of the following methods in a test suite:
// Run once when the suite starts running.
func (s *FtpClientTestSuite) SetUpSuite(c *C) {
}

//Run before each test or benchmark starts running.
func (s *FtpClientTestSuite) SetUpTest(c *C) {

	var FtpsClientParam_X FtpsClientParam

	FtpsClientParam_X.Id_U32 = CONID
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

	GL_FtpsClientPtr_X = NewFtpsClient(&FtpsClientParam_X)
	if GL_FtpsClientPtr_X == nil {
		c.Fatalf("Unable to create client !")
	}
	Err := GL_FtpsClientPtr_X.Connect()
	if Err != nil {
		c.Fatalf("Connect error: %v\n", Err)
	}
	if GL_FtpsClientPtr_X.FtpsParam_X.Id_U32 != CONID {
		c.Fatalf("Bad param %v instead of %v\n", GL_FtpsClientPtr_X.FtpsParam_X.Id_U32, CONID)
	}
	Err = GL_FtpsClientPtr_X.DeleteFile(HOST_PATH)
}

//Run after each test or benchmark runs.
func (s *FtpClientTestSuite) TearDownTest(c *C) {
	if GL_FtpsClientPtr_X == nil {
		c.Fatalf("Invalid client (see SetUpTest) !")
	}

	Err := GL_FtpsClientPtr_X.Disconnect()
	if Err != nil {
		c.Fatalf("Disconnect error: %v\n", Err)
	}
}

//Run once after all tests or benchmarks have finished running.
func (s *FtpClientTestSuite) TearDownSuite(c *C) {

}

const (
	CONID     = 123
	MKDIRNAME = "bha"
	IOFILE    = "ftpsclient.go"
)

func (s *FtpClientTestSuite) TestDirectory(c *C) {
	Directory_S, Err := GL_FtpsClientPtr_X.GetWorkingDirectory()
	if Err != nil {
		c.Fatalf("GetWorkingDirectory error: %v\n", Err)
	}
	if Directory_S != GL_FtpsClientPtr_X.FtpsParam_X.InitialDirectory_S {
		c.Fatalf("Bad directory %v instead of %v\n", Directory_S, GL_FtpsClientPtr_X.FtpsParam_X.InitialDirectory_S)
	}
	//Just to be sure !
	Err = GL_FtpsClientPtr_X.RemoveDirectory(MKDIRNAME)
	//	checkTestError(c, Err, "RemoveDirectory error: %v\n", Err)

	Err = GL_FtpsClientPtr_X.MakeDirectory(MKDIRNAME)
	if Err != nil {
		c.Fatalf("MakeDirectory error: %v\n", Err)
	}

	Err = GL_FtpsClientPtr_X.ChangeWorkingDirectory(MKDIRNAME)
	if Err != nil {
		c.Fatalf("ChangeWorkingDirectory error: %v\n", Err)
	}

	Directory_S, Err = GL_FtpsClientPtr_X.GetWorkingDirectory()
	if Err != nil {
		c.Fatalf("GetWorkingDirectory error: %v\n", Err)
	}
	NewDir_S := GL_FtpsClientPtr_X.FtpsParam_X.InitialDirectory_S + "/" + MKDIRNAME
	if Directory_S != NewDir_S {
		c.Fatalf("Bad directory %v instead of %v\n", Directory_S, NewDir_S)
	}

	Err = GL_FtpsClientPtr_X.ChangeWorkingDirectory("..")
	if Err != nil {
		c.Fatalf("ChangeWorkingDirectory error: %v\n", Err)
	}
	Err = GL_FtpsClientPtr_X.RemoveDirectory(MKDIRNAME)
	if Err != nil {
		c.Fatalf("RemoveDirectory error: %v\n", Err)
	}

}

func (s *FtpClientTestSuite) TestDisconnect(c *C) {
	Err := GL_FtpsClientPtr_X.Disconnect()
	if Err != nil {
		c.Fatalf("Disconnect error: %v\n", Err)
	}
	s.SetUpTest(c)
}

func (s *FtpClientTestSuite) TestUpload(c *C) {

	Err := uploadFile(LOCAL_PATH, HOST_PATH)
	if Err != nil {
		c.Fatalf("Upload error: %v\n", Err)
	}
}

func (s *FtpClientTestSuite) TestDownload(c *C) {
	Err := uploadFile(LOCAL_PATH, HOST_PATH)
	if Err != nil {
		c.Fatalf("Upload error: %v\n", Err)
	} else {
		Err := GL_FtpsClientPtr_X.RetrieveFile(HOST_PATH, LOCAL_PATH_RETR)
		if Err != nil {
			c.Fatalf("RetrieveFile error: %v\n", Err)
		}
	}
}

func (s *FtpClientTestSuite) TestFileList(c *C) {
	Err := uploadFile(LOCAL_PATH, HOST_PATH)
	if Err != nil {
		c.Fatalf("Upload error: %v\n", Err)
	} else {
		//		pDirEntry_X, Err := GL_FtpsClientPtr_X.List()
		_, Err := GL_FtpsClientPtr_X.List()
		if Err != nil {
			c.Fatalf("List error: %v\n", Err)
		}
		//			for _, DirEntry_X := range pDirEntry_X {
		//						log.Println(fmt.Sprintf("(%d): %s.%s %d bytes %s", DirEntry_X.Type_E, DirEntry_X.Name_S, DirEntry_X.Ext_S, DirEntry_X.Size_U64, DirEntry_X.Time_X))
		//		}

	}
}

func (s *FtpClientTestSuite) TestFileDelete(c *C) {
	Err := uploadFile(LOCAL_PATH, HOST_PATH)
	if Err != nil {
		c.Fatalf("Upload error: %v\n", Err)
	} else {
		Err := GL_FtpsClientPtr_X.DeleteFile(HOST_PATH)
		if Err != nil {
			c.Fatalf("DeleteFile error: %v\n", Err)
		}
	}
}

func (s *FtpClientTestSuite) TestCtrlCmd(c *C) {
	ReplyCode_i, ReplyMessage_S, Err := GL_FtpsClientPtr_X.SendFtpCtrlCommand("FEAT", 211)

	if Err != nil {
		c.Fatalf("SendFtpCtrlCommand error: %v %d %s\n", Err, ReplyCode_i, ReplyMessage_S)
	}
}

func (s *FtpClientTestSuite) TestDataChannel(c *C) {
	var DataArray_U8 [0x1000]uint8

	Err := uploadFile(LOCAL_PATH, HOST_PATH)
	if Err != nil {
		c.Fatalf("Upload error: %v\n", Err)
	}

	ReplyCode_i, ReplyMessage_S, Err := GL_FtpsClientPtr_X.OpenFtpDataChannel("LIST", 150)
	if Err != nil {
		c.Fatalf("OpenFtpDataChannel error: %v %d %s\n", Err, ReplyCode_i, ReplyMessage_S)
	}
	_, _, NbRead_i, Err := GL_FtpsClientPtr_X.ReadFtpDataChannel(true, DataArray_U8[:])
	if Err != nil {
		c.Fatalf("ReadFtpDataChannel error: %v\n", Err, NbRead_i)
	}
	ReplyCode_i, ReplyMessage_S, Err = GL_FtpsClientPtr_X.CloseFtpDataChannel()
	if Err != nil {
		c.Fatalf("CloseFtpDataChannel error: %v %d %s\n", Err, ReplyCode_i, ReplyMessage_S)
	}

}
func uploadFile(_LocalPath_S, _HostPath_S string) error {

	pData_U8, rRts := ioutil.ReadFile(_LocalPath_S)
	if rRts == nil {
		rRts = GL_FtpsClientPtr_X.StoreFile(_HostPath_S, pData_U8)
	}
	return rRts
}
