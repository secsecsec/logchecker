// Copyright (c) 2015, Alexander Zaytsev. All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

// LogChecker package is a simple library to check a list of logs files
// and send notification about their abnormal activities.
//
// Error logger is activated by default,
// use DebugMode method to turn on debug mode:
//
//     DebugMode(true)
//
// Initialization from file:
//
//     logger := logchecker.New()
//     if err := logchecker.InitConfig(logger, "filiename"); err != nil {
//         // error detected
//     }
//
// Manually initialization of setting to send emails:
//
//     logger := logchecker.New()
//     logger.Cfg.Sender := map[string]string{
//      "user": "user@host.com",
//      "password": "password",
//      "host": "smtp.host.com",
//      "addr": "smtp.host.com:25",
//     }
//
package logchecker

import (
    "os"
    "log"
    "fmt"
    "strings"
    "io/ioutil"
    "encoding/json"
    "path/filepath"
)

var (
    LoggerError *log.Logger = log.New(os.Stderr, "LogChecker ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
    LoggerDebug *log.Logger = log.New(ioutil.Discard, "LogChecker DEBUG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
)

type File struct {
    Log string      `json:"file"`
    Delay uint      `json:"delay"`
    Pattern string  `json:"pattern"`
    Boundary uint   `json:"boundary"`
    Increase bool `json:"increase"`
    Emails []string `json:"emails"`
    Limits []uint   `json:"limits"`
}

type Service struct {
    Name string   `json:"name"`
    Files []File  `json:"files"`
}

type Config struct {
    Path string
    Sender map[string]string  `json:"sender"`
    Observed []Service        `json:"observed"`
}

func (cfg Config) String() string {
    services := make([]string, len(cfg.Observed))
    for i, service := range cfg.Observed {
        // services[i] = fmt.Sprintf("%v", service.Name)
        files := make([]string, len(service.Files))
        for j, file := range service.Files {
            files[j] = fmt.Sprintf("File: %v; Delay: %v; Pattern: %v; Boundary: %v; Increase: %v; Emails: %v; Limits: %v", file.Log, file.Delay, file.Pattern, file.Boundary, file.Increase, file.Emails, file.Limits)
        }
        services[i] = fmt.Sprintf("%v\n\t%v", service.Name, strings.Join(files, "\n\t"))
    }
    return fmt.Sprintf("Config: %v\n sender: %v\n---\n%v", cfg.Path, cfg.Sender, strings.Join(services, "\n---\n"))
}

type LogChecker struct {
    Name string
    Cfg Config
}

// Initialization of Logger handlers
func DebugMode(debugmode bool) {
    debugHandle := ioutil.Discard
    if debugmode {
        debugHandle = os.Stdout
    }
    LoggerDebug = log.New(debugHandle, "LogChecker DEBUG: ",
        log.Ldate|log.Lmicroseconds|log.Lshortfile)
}

// New LogChecker object
func New() *LogChecker {
    return &LogChecker{}
}

// It validates file name, converts its path from relative to absolute
// using current directory address.
func FilePath(name string) (string, error) {
    var (
        fullpath string
        err error
    )
    fullpath = strings.Trim(name, " ")
    if len(fullpath) < 1 {
        return fullpath, fmt.Errorf("Empty file name")
    }
    fullpath, err = filepath.Abs(fullpath)
    if err != nil {
        return fullpath, err
    }
    _, err = os.Stat(fullpath);
    return fullpath, err
}

// Initializes configuration from a file.
func InitConfig(logger *LogChecker, name string) error {
    path, err := FilePath(name)
    if err != nil {
        LoggerError.Println("Can't check config file")
        return err
    }
    logger.Cfg.Path = path
    jsondata, err := ioutil.ReadFile(path)
    if err != nil {
        LoggerError.Println("Can't read config file")
    }
    if err = json.Unmarshal(jsondata, &logger.Cfg); err != nil {
        LoggerError.Println("Can't parse config file")
    }
    return err
}
