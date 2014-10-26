ftpsclient
==========

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

