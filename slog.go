package slog

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	VERSION string = "0.0.1"
)

type LEVEL int32

var logLevel LEVEL = 1
var maxFileSize int64
var maxFileCount int32
var dailyRolling bool = true
var consoleAppender bool = true
var RollingFile bool = false
var logObj *_FILE

const DATEFORMAT = "2006-01-02"

type UNIT int64

const (
	_       = iota
	KB UNIT = 1 << (iota * 10)
	MB
	GB
	TB
)

const (
	BLACK      = "30"
	RED        = "31" //fatal default color
	GREEN      = "32" //info default color
	YELLO      = "33" //warn default color
	BLUE       = "34" //debug default color
	PURPLE_RED = "35" //error default color
	CYAN_BLUE  = "36"
	WHITE      = "37"

	LOG_START = "\033[1;0;"
	LOG_END   = "\033[0m"
)

const (
	ALL LEVEL = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
	OFF
)

type _FILE struct {
	dir      string
	filename string
	_suffix  int
	isCover  bool
	_date    *time.Time
	mu       *sync.RWMutex
	logfile  *os.File
	lg       *log.Logger
}

func SetConsole(isConsole bool) {
	consoleAppender = isConsole
}

func SetLevel(_level LEVEL) {
	logLevel = _level
}

//修改以时间为结尾命名
func SetRollingFile(fileDir, fileName string, maxNumber int32, maxSize int64, _unit UNIT) {
	maxFileCount = maxNumber
	maxFileSize = maxSize * int64(_unit)
	RollingFile = true
	dailyRolling = false
	logObj = &_FILE{dir: fileDir, filename: fileName, isCover: false, mu: new(sync.RWMutex)}
	logObj.mu.Lock()
	defer logObj.mu.Unlock()
	for i := 1; i <= int(maxNumber); i++ {
		if isExist(fileDir + "/" + fileName + "." + strconv.Itoa(i)) {
			logObj._suffix = i
		} else {
			break
		}
	}
	if !logObj.isMustRename() {
		logObj.logfile, _ = os.OpenFile(fileDir+"/"+fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		logObj.lg = log.New(logObj.logfile, "\n", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	} else {
		logObj.rename()
	}
	go fileMonitor()
}

func SetRollingDaily(fileDir, fileName string) {
	RollingFile = false
	dailyRolling = true
	t, _ := time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))
	logObj = &_FILE{dir: fileDir, filename: fileName, _date: &t, isCover: false, mu: new(sync.RWMutex)}
	logObj.mu.Lock()
	defer logObj.mu.Unlock()

	if !logObj.isMustRename() {
		logObj.logfile, _ = os.OpenFile(fileDir+"/"+fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		logObj.lg = log.New(logObj.logfile, "\n", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	} else {
		logObj.rename()
	}
}

func console(sflag string, s ...interface{}) {
	if consoleAppender {
		_, file, line, _ := runtime.Caller(2)
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		var data string
		if runtime.GOOS == "windows" {
			log.Println(file+":"+strconv.Itoa(line), s)
		} else {
			sfile := LOG_START + BLUE + "m" + file + LOG_END + ":" + LOG_START + YELLO + "m" + strconv.Itoa(line) + LOG_END + "[" + LOG_START + BLUE + "m" + sflag + LOG_END + "]=>"
			val := fmt.Sprintln(s)
			val = strings.Replace(val, "\n", "", -1)
			val = strings.Replace(val, "[[", "", -1)
			val = strings.Replace(val, "]]", "", -1)
			fhead := fmt.Sprintf(sfile)
			switch sflag {
			case "debug":
				data = fhead + LOG_START + WHITE + "m" + val + LOG_END
			case "info":
				data = fhead + LOG_START + WHITE + "m" + val + LOG_END
			case "warn":
				data = fhead + LOG_START + YELLO + "m" + val + LOG_END
			case "error":
				data = fhead + LOG_START + PURPLE_RED + "m" + val + LOG_END
			case "fatal":
				data = fhead + LOG_START + RED + "m" + val + LOG_END
			default:
				data = fhead + LOG_START + WHITE + "m" + val + LOG_END
			}
			log.Println(data)
		}
	}
}

func catchError() {
	if err := recover(); err != nil {
		log.Println("err", err)
	}
}

func Debug(v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()

	if logLevel <= DEBUG {
		logObj.lg.Output(2, fmt.Sprintln("debug", v))
		console("debug", v)
	}
}
func Info(v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()
	if logLevel <= INFO {
		logObj.lg.Output(2, fmt.Sprintln("info", v))
		console("info", v)
	}
}
func Warn(v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()
	if logLevel <= WARN {
		logObj.lg.Output(2, fmt.Sprintln("warn", v))
		console("warn", v)
	}
}
func Error(v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()
	if logLevel <= ERROR {
		logObj.lg.Output(2, fmt.Sprintln("error", v))
		console("error", v)
	}
}
func Fatal(v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()
	if logLevel <= FATAL {
		logObj.lg.Output(2, fmt.Sprintln("fatal", v))
		console("fatal", v)
	}
}

func (f *_FILE) isMustRename() bool {
	if dailyRolling {
		t, _ := time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))
		if t.After(*f._date) {
			return true
		}
	} else {
		if maxFileCount > 1 {
			if fileSize(f.dir+"/"+f.filename) >= maxFileSize {
				return true
			}
		}
	}
	return false
}

func (f *_FILE) rename() {
	if dailyRolling {
		fn := f.dir + "/" + f.filename + "." + f._date.Format(DATEFORMAT)
		if !isExist(fn) && f.isMustRename() {
			if f.logfile != nil {
				f.logfile.Close()
			}
			err := os.Rename(f.dir+"/"+f.filename, fn)
			if err != nil {
				f.lg.Println("rename err", err.Error())
			}
			t, _ := time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))
			f._date = &t
			f.logfile, _ = os.Create(f.dir + "/" + f.filename)
			f.lg = log.New(logObj.logfile, "\n", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
		}
	} else {
		f.coverNextOne()
	}
}

func (f *_FILE) nextSuffix() int {
	return int(f._suffix%int(maxFileCount) + 1)
}

func (f *_FILE) coverNextOne() {
	f._suffix = f.nextSuffix()
	if f.logfile != nil {
		f.logfile.Close()
	}
	if isExist(f.dir + "/" + f.filename + "." + strconv.Itoa(int(f._suffix))) {
		os.Remove(f.dir + "/" + f.filename + "." + strconv.Itoa(int(f._suffix)))
	}
	os.Rename(f.dir+"/"+f.filename, f.dir+"/"+f.filename+"."+strconv.Itoa(int(f._suffix)))
	f.logfile, _ = os.Create(f.dir + "/" + f.filename)
	f.lg = log.New(logObj.logfile, "\n", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
}

func fileSize(file string) int64 {
	fmt.Println("fileSize", file)
	f, e := os.Stat(file)
	if e != nil {
		fmt.Println(e.Error())
		return 0
	}
	return f.Size()
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func fileMonitor() {
	timer := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timer.C:
			fileCheck()
		}
	}
}

func fileCheck() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	if logObj != nil && logObj.isMustRename() {
		logObj.mu.Lock()
		defer logObj.mu.Unlock()
		logObj.rename()
	}
}
