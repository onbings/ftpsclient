// Copyright 2014 OnBings. All rights reserved.
// Use of this source code is governed by a APACHE-style
// license that can be found in the LICENSE file.
12kk
/*
	This module implements the 'ftpsclient' package unit test.
	To run theses test please install Filezilla Ftp server (https://filezilla-project.org/download.php?type=server)
	and create a ftp server config with a user name 'mc' with pasword 'a'. The ftp server should exposes a 'Seq'
	directory.

	These unit tests use the 'gocheck' golang test package (https://labix.org/gocheck)

*/
package ftpsclient

import (
	"fmt"
	. "gopkg.in/check.v1"
	_ "io/ioutil"
	_ "log"
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
	fmt.Printf("SetUpSuite\n")

}

//Run before each test or benchmark starts running.
func (s *FtpClientTestSuite) SetUpTest(c *C) {
	fmt.Printf("SetUpTest")

}

//Run after each test or benchmark runs.
func (s *FtpClientTestSuite) TearDownTest(c *C) {
	fmt.Printf("TearDownTest")

}

//Run once after all tests or benchmarks have finished running.
func (s *FtpClientTestSuite) TearDownSuite(c *C) {
	fmt.Printf("TearDownSuite")

}

func (s *FtpClientTestSuite) TestHelloWorld(c *C) {
	c.Assert(42, Equals, "42")
	c.Check(42, Equals, 42)
}

const (
	CONID     = 123
	MKDIRNAME = "bha"
	IOFILE    = "ftpsclient.go"
)

func checkTestError(c *C, _Err error, _Format_S string, args ...interface{}) {
	if _Err != nil {
		c.Fatalf(_Format_S, args...)
	}
}
func connect(c *C) {
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
	checkTestError(c, Err, "Connect error: %v\n", Err)
	if GL_FtpsClientPtr_X.FtpsParam_X.Id_U32 != CONID {
		c.Fatalf("Bad param %v instead of %v\n", GL_FtpsClientPtr_X.FtpsParam_X.Id_U32, CONID)
	}
}
func TestDirectory(c *C) {
	Directory_S, Err := GL_FtpsClientPtr_X.GetWorkingDirectory()
	checkTestError(c, Err, "GetWorkingDirectory error: %v\n", Err)
	if Directory_S != GL_FtpsClientPtr_X.FtpsParam_X.InitialDirectory_S {
		c.Fatalf("Bad directory %v instead of %v\n", Directory_S, GL_FtpsClientPtr_X.FtpsParam_X.InitialDirectory_S)
	}
	Err = GL_FtpsClientPtr_X.MakeDirectory(MKDIRNAME)
	checkTestError(c, Err, "MakeDirectory error: %v\n", Err)

	Err = GL_FtpsClientPtr_X.ChangeWorkingDirectory(MKDIRNAME)
	checkTestError(c, Err, "ChangeWorkingDirectory error: %v\n", Err)

	Directory_S, Err = GL_FtpsClientPtr_X.GetWorkingDirectory()
	checkTestError(c, Err, "GetWorkingDirectory error: %v\n", Err)
	NewDir_S := GL_FtpsClientPtr_X.FtpsParam_X.InitialDirectory_S + "/" + MKDIRNAME
	if Directory_S != NewDir_S {
		c.Fatalf("Bad directory %v instead of %v\n", Directory_S, NewDir_S)
	}
}

func TestDisconnect(c *C) {
	Err := GL_FtpsClientPtr_X.Disconnect()
	checkTestError(c, Err, "Disconnect error: %v\n", Err)
	/*
		Err = GL_FtpsClientPtr_X.MakeDirectory("bha")
		if Err != nil {
			panic(Err)
		}

		Err = GL_FtpsClientPtr_X.ChangeWorkingDirectory("Seq")
		if Err != nil {
			panic(Err)
		}

		Directory_S, Err = GL_FtpsClientPtr_X.GetWorkingDirectory()
		if Err != nil {
			panic(Err)
		}
		log.Printf("Current working Directory_S: %s", Directory_S)

		Err = GL_FtpsClientPtr_X.ChangeWorkingDirectory("..")
		if Err != nil {
			panic(Err)
		}

		Directory_S, Err = GL_FtpsClientPtr_X.GetWorkingDirectory()
		if Err != nil {
			panic(Err)
		}
		log.Printf("Current working Directory_S: %s", Directory_S)

		Err = GL_FtpsClientPtr_X.RemoveDirectory("bha")
		if Err != nil {
			panic(Err)
		}

		Directory_S, Err = GL_FtpsClientPtr_X.GetWorkingDirectory()
		if Err != nil {
			panic(Err)
		}
		log.Printf("Current working Directory_S: %s", Directory_S)

		pData_U8, Err := ioutil.ReadFile("ftpsclient.go")
		if Err != nil {
			panic(Err)
		}
		Err = GL_FtpsClientPtr_X.StoreFile("test2", pData_U8)
		if Err != nil {
			panic(Err)
		}

		Err = GL_FtpsClientPtr_X.RetrieveFile("test2", "test3")
		if Err != nil {
			panic(Err)
		}

		pDirEntry_X, Err := GL_FtpsClientPtr_X.List()
		if Err != nil {
			panic(Err)
		}
		for _, DirEntry_X := range pDirEntry_X {
			log.Println(DirEntry_X)
		}

		Err = GL_FtpsClientPtr_X.DeleteFile("test2")
		if Err != nil {
			panic(Err)
		}

		Err = GL_FtpsClientPtr_X.Disconnect()
		if Err != nil {
			panic(Err)
		}
	*/
}
