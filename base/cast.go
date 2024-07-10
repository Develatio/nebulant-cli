// MIT License
//
// Copyright (C) 2020  Develatio Technologies S.L.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package base

const SilentLevel = 100

// CriticalLevel const
const CriticalLevel = 50

// ErrorLevel const
const ErrorLevel = 40

// WarningLevel const
const WarningLevel = 30

// InfoLevel const
const InfoLevel = 20

// DebugLevel const
const DebugLevel = 10

// DebugLevel const
const ParanoicDebugLevel = 5

// NotsetLevel const
const NotsetLevel = 0

// ILogger interface
type ILogger interface {
	LogCritical(s string)
	LogErr(s string)
	ByteLogErr(b []byte)
	LogWarn(s string)
	LogInfo(s string)
	ByteLogInfo(b []byte)
	LogDebug(s string)
	Duplicate() ILogger
	SetActionID(ai string)
	SetActionName(ai string)
	SetThreadID(ti string)
}
