// Copyright 2020 the go-etl Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spi

import "github.com/Breeze0806/go-etl/datax/common/spi/reader"

//Reader 读取器
type Reader interface {
	Job() reader.Job   //获取读取工作,一般不能为空
	Task() reader.Task //获取读取任务,一般不能为空
}
