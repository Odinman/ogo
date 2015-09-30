package logs

import (
	"encoding/json"
	//"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// AccessLogWriter implements LoggerInterface.
// It writes messages by lines limit, file size limit, or time frequency.
type AccessLogWriter struct {
	*log.Logger
	mw *AccessMuxWriter
	// The opened file
	Filename string `json:"filename"`

	// Rotate at size
	Maxsize         int `json:"maxsize"`
	maxsize_cursize int

	// Rotate daily
	Daily          bool  `json:"daily"`
	Maxdays        int64 `json:"maxdays`
	daily_opendate int

	Rotate    bool `json:"rotate"`
	MaxRotate int  `json:"maxrotate"`

	startLock sync.Mutex // Only one log can write to the file

	Level int `json:"level"`
}

// an *os.File writer with locker.
type AccessMuxWriter struct {
	sync.Mutex
	fd *os.File
}

// write to os.File.
func (l *AccessMuxWriter) Write(b []byte) (int, error) {
	l.Lock()
	defer l.Unlock()
	return l.fd.Write(b)
}

// set os.File in writer.
func (l *AccessMuxWriter) SetFd(fd *os.File) {
	if l.fd != nil {
		l.fd.Close()
	}
	l.fd = fd
}

// create a AccessLogWriter returning as LoggerInterface.
func NewAccessWriter() LoggerInterface {
	w := &AccessLogWriter{
		Filename:  "logs/access.log",
		Maxsize:   1 << 29, //512 MB
		Daily:     false,
		Maxdays:   30, //旧日志保存30天
		Rotate:    true,
		MaxRotate: 10, //保持10个旧文件(与3天条件共存)
		Level:     LevelTrace,
	}
	// use AccessMuxWriter instead direct use os.File for lock write when rotate
	w.mw = new(AccessMuxWriter)
	// set AccessMuxWriter as Logger's io.Writer
	//w.Logger = log.New(w.mw, "", log.Ldate|log.Ltime)
	w.Logger = log.New(w.mw, "", 0)
	return w
}

// Init file logger with json config.
// jsonconfig like:
//	{
//	"filename":"logs/access.log",
//	"maxsize":1<<30,
//	"daily":true,
//	"maxdays":15,
//	"rotate":true
//	}
func (w *AccessLogWriter) Init(jsonconfig string) error {
	err := json.Unmarshal([]byte(jsonconfig), w)
	if err != nil {
		return err
	}
	//if len(w.Filename) == 0 {
	//    return errors.New("jsonconfig must have filename")
	//}
	err = w.startLogger()
	return err
}

// start file logger. create log file and set to locker-inside file writer.
func (w *AccessLogWriter) startLogger() error {
	fd, err := w.createLogFile()
	if err != nil {
		return err
	}
	w.mw.SetFd(fd)
	err = w.initFd()
	if err != nil {
		return err
	}
	return nil
}

func (w *AccessLogWriter) docheck(size int) {
	w.startLock.Lock()
	defer w.startLock.Unlock()
	if w.Rotate && ((w.Maxsize > 0 && w.maxsize_cursize >= w.Maxsize) || (w.Daily && time.Now().Day() != w.daily_opendate)) {
		if err := w.DoRotate(); err != nil {
			fmt.Fprintf(os.Stderr, "AccessLogWriter(%q): %s\n", w.Filename, err)
			return
		}
	}
	w.maxsize_cursize += size
}

// write logger message into file.
func (w *AccessLogWriter) WriteMsg(msg string, level int) error {
	if level < w.Level {
		return nil
	}
	w.docheck(len(msg))
	w.Logger.Println(msg)
	return nil
}

func (w *AccessLogWriter) createLogFile() (*os.File, error) {
	dir := filepath.Dir(w.Filename)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			//mkdir
			if err := os.Mkdir(dir, 0755); err != nil {
				return nil, err
			}
		}
	}
	// Open the log file
	fd, err := os.OpenFile(w.Filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	return fd, err
}

func (w *AccessLogWriter) initFd() error {
	fd := w.mw.fd
	finfo, err := fd.Stat()
	if err != nil {
		return fmt.Errorf("get stat err: %s\n", err)
	}
	w.maxsize_cursize = int(finfo.Size())
	w.daily_opendate = time.Now().Day()
	return nil
}

/* {{{ func fileRotateRename(filename string, numbers ...int) error
 * 轮询移动文件,直接用数字后缀0-base
 */
func fileRotateRename(filename string, numbers ...int) error {
	var source, target string
	var sn, max int
	max = 10 //默认10个
	if len(numbers) > 0 {
		sn = numbers[0]
	}
	if len(numbers) > 1 {
		max = numbers[1]
	}
	target = fmt.Sprint(filename, ".", sn)
	if sn > 0 {
		source = fmt.Sprint(filename, ".", sn-1)
	} else {
		source = filename
	}
	if _, err := os.Lstat(target); err == nil && (max <= 0 || sn+1 < max) {
		//文件存在,并且没有达到最大rotate限制
		if err := fileRotateRename(filename, sn+1, max); err != nil { //递归腾地儿
			return err
		}
	}
	return os.Rename(source, target)
}

/* }}} */

// DoRotate means it need to write file in new file.
// new file name like xx.log.2013-01-01.2
func (w *AccessLogWriter) DoRotate() error {
	_, err := os.Lstat(w.Filename)
	if err == nil { // file exists

		// block Logger's io.Writer
		w.mw.Lock()
		defer w.mw.Unlock()

		fd := w.mw.fd
		fd.Close()

		// close fd before rename
		// Rename the file to its newfound home
		//err = os.Rename(w.Filename, fname)
		//if err != nil {
		if err := fileRotateRename(w.Filename, 0, 10); err != nil {
			return fmt.Errorf("Rotate: %s\n", err)
		}

		// re-start logger
		err = w.startLogger()
		if err != nil {
			return fmt.Errorf("Rotate StartLogger: %s\n", err)
		}

		go w.deleteOldLog()
	}

	return nil
}

func (w *AccessLogWriter) deleteOldLog() {
	dir := filepath.Dir(w.Filename)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && info.ModTime().Unix() < (time.Now().Unix()-60*60*24*w.Maxdays) {
			if strings.HasPrefix(filepath.Base(path), filepath.Base(w.Filename)) {
				os.Remove(path)
			}
		}
		return nil
	})
}

// destroy file logger, close file writer.
func (w *AccessLogWriter) Destroy() {
	w.mw.fd.Close()
}

// flush file logger.
// there are no buffering messages in file logger in memory.
// flush file means sync file from disk.
func (w *AccessLogWriter) Flush() {
	w.mw.fd.Sync()
}

// log access message.
func (ol *OLogger) Access(msg string) {
	lm := new(logMsg)
	lm.level = LevelCritical //最高级别
	lm.msg = msg
	ol.msg <- lm
}

func init() {
	Register("access", NewAccessWriter)
}
