// Copyright (c) 2015, Alexander Zaytsev. All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

// LogChecker testing methods
//
package logchecker

import (
    "os"
    // "fmt"
    "testing"
    "strings"
    "io/ioutil"
    "path/filepath"
)

func buildDir() string {
    path := os.Getenv("TRAVIS_BUILD_DIR")
    if len(path) > 0 {
        return path
    }
    return filepath.Join(os.Getenv("GOPATH"), "src/github.com/z0rr0/logchecker")
}

func createFile(name string, mode int) (*os.File, error) {
    file, err := os.Create(name)
    if err != nil {
        return nil, err
    }
    return file, file.Chmod(os.FileMode(mode))
}

func prepareConfig(from, to string, replace map[string]string) error {
    // newfiles := [3]string("error.log", "access.log", "syslog")
    data, err := ioutil.ReadFile(from)
    if err != nil {
        return err
    }
    strinfo := string(data)
    for k, v := range replace {
        strinfo = strings.Replace(strinfo, k, v, 1)
    }
    return ioutil.WriteFile(to, []byte(strinfo), os.FileMode(0666))
}


func TestDebugMode(t *testing.T) {
    if (LoggerError == nil) || (LoggerDebug == nil) {
        t.Errorf("incorrect references")
    }
    DebugMode(false)
    if (LoggerError.Prefix() != "LogChecker ERROR: ") || (LoggerDebug.Prefix() != "LogChecker DEBUG: ") {
        t.Errorf("incorrect loggers settings")
    }
    DebugMode(true)
    if (LoggerError.Flags() != 19) || (LoggerDebug.Flags() != 21) {
        t.Errorf("incorrect loggers settings")
    }
}

func TestNew(t *testing.T) {
    logger := New()
    if logger == nil {
        t.Errorf("incorrect reference")
    }
    serv := Service{}
    if err := logger.AddService(&serv); err == nil {
        t.Errorf("incorrect response for empty Service: %v\n", err)
    }
    serv.Name = "TestSrv"
    if logger.HasService(&serv, true) {
        t.Errorf("incorrect response")
    }
    if err := logger.AddService(&serv); err != nil {
        t.Errorf("incorrect response: %v\n", err)
    }
    if !logger.HasService(&serv, true) {
        t.Errorf("incorrect response")
    }
    if err := logger.AddService(&serv); err == nil {
        t.Errorf("incorrect response: %v\n", err)
    }
    if err := logger.Validate(); err == nil {
        t.Errorf("incorrect response: %v\n", err)
    }
    logger.Cfg.Sender = map[string]string{
        "user": "user@host.com",
        "password": "password",
        "host": "smtp.host.com",
        // "addr": "smtp.host.com:25",
    }
    if err := logger.Validate(); err == nil {
        t.Errorf("incorrect response: %v\n", err)
    }
    logger.Cfg.Sender = map[string]string{
        "user": "user@host.com",
        "password": "password",
        "host": "smtp.host.com",
        "addr": "",
    }
    if err := logger.Validate(); err == nil {
        t.Errorf("incorrect response: %v\n", err)
    }
    logger.Cfg.Sender = map[string]string{
        "user": "user@host.com",
        "password": "password",
        "host": "smtp.host.com",
        "addr": "smtp.host.com:25",
    }
    if err := logger.Validate(); err == nil {
        t.Errorf("incorrect response: %v\n", err)
    }
    logger.Cfg.Storage = "memory"
    if err := logger.Validate(); err != nil {
        t.Errorf("incorrect response: %v\n", err)
    }
    if logger.Backend.GetName() != "Memory" {
        t.Errorf("incorrect backend name: %v\n", logger.Backend.GetName())
    }
}

func TestFilePath(t *testing.T) {
    if _, err := FilePath("invalid_name"); err == nil {
        t.Errorf("incorrect response")
    }
    if _, err := FilePath(""); err == nil {
        t.Errorf("incorrect response")
    }
    // if _, err := FilePath("unknown"); err == nil {
    //     t.Errorf("incorrect response")
    // }
    realfile := filepath.Join(buildDir(), "config.example.json")
    if path, err := FilePath(realfile); err != nil {
        t.Errorf("incorrect response, the file should exist")
    } else {
        if path != realfile {
            t.Errorf("ivalid paths")
        }
    }
}

func TestInitConfig(t *testing.T) {
    testdir := buildDir()
    newvalues := map[string]string{
        "/var/log/nginx/error.log": filepath.Join(testdir, "error.log"),
        "/var/log/nginx/access.log": filepath.Join(testdir, "access.log"),
        "/var/log/syslog": filepath.Join(testdir, "syslog"),
    }
    oldexample := filepath.Join(testdir, "config.example.json")
    example := filepath.Join(testdir, "config.new.json")
    if err := prepareConfig(oldexample, example, newvalues); err != nil {
        t.Errorf("can't prepare test config file [%v]", err)
    }
    defer os.Remove(example)
    for _, v := range newvalues {
        if _, err := createFile(v, 0666); err != nil {
            t.Errorf("test file preparation error [%v]", err)
        }
        defer os.Remove(v)
    }
    logger := New()
    if err := InitConfig(logger, example); err != nil {
        t.Errorf("error during InitConfig [%v]: %v", example, err)
    }

    if len(logger.Cfg.Observed) > 1 {
        logger.Cfg.Observed[1].Name = "Nginx"
        if err := logger.Validate(); err == nil {
            t.Errorf("wrong validation [%v]", err)
        }
    }
    if l := len(logger.Cfg.String()); l <= 0 {
        t.Errorf("config should be initiated [%v]", l)
    }
    if err := InitConfig(logger, "invalid_name"); err == nil {
        t.Errorf("need error during InitConfig")
    }

    testfile := filepath.Join(testdir, "testfile.json")
    f, err := createFile(testfile, 0200);
    if err != nil {
        t.Errorf("%v", err)
    }
    defer os.Remove(testfile)

    if err := InitConfig(logger, testfile); err == nil {
        t.Errorf("need permissions error during InitConfig")
    }

    f.Chmod(0600)
    if err := InitConfig(logger, "/etc/passwd"); err == nil {
        t.Errorf("need json error during InitConfig")
    }
}
