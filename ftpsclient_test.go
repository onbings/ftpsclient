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
	//	"fmt"
	. "gopkg.in/check.v1"
	"io/ioutil"
	"log"
	"testing"
)

//-- GoCheck specific initialization --------------------------------------

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type FtpClientTestSuite struct{}

var _ = Suite(&FtpClientTestSuite{})

//-- GoCheck specific initialization --------------------------------------

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
	FtpsClientParam_X.Debug_B = true
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

/*
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
*/
func (s *FtpClientTestSuite) TestFileUpload(c *C) {

	pData_U8, Err := ioutil.ReadFile("ftpsclient.go")
	if Err != nil {
		c.Fatalf("ReadFile error: %v\n", Err)
	}
	Err = GL_FtpsClientPtr_X.StoreFile("test2", pData_U8)
	if Err != nil {
		c.Fatalf("StoreFile error: %v\n", Err)
	}

}

func (s *FtpClientTestSuite) TestFileDownload(c *C) {
	Err := GL_FtpsClientPtr_X.RetrieveFile("test2", "test3")
	if Err != nil {
		c.Fatalf("RetrieveFile error: %v\n", Err)
	}
}
func (s *FtpClientTestSuite) TestFileList(c *C) {
	pDirEntry_X, Err := GL_FtpsClientPtr_X.List()
	if Err != nil {
		c.Fatalf("List error: %v\n", Err)
	}
	for _, DirEntry_X := range pDirEntry_X {
		log.Println(DirEntry_X)
	}
}

func (s *FtpClientTestSuite) TestFileDelete(c *C) {
	Err := GL_FtpsClientPtr_X.DeleteFile("test2")
	if Err != nil {
		c.Fatalf("DeleteFile error: %v\n", Err)
	}

}
