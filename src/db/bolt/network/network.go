/*******************************************************************************
 * Copyright 2019 Samsung Electronics All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 *******************************************************************************/
package network

import (
	"encoding/json"

	"common/errors"
	bolt "db/bolt/wrapper"
)

const bucketName = "network"

type Addr struct {
	Wired map[string]string `json:"wired"`
	Wireless map[string]string `json:"wireless"`
}

type NetworkInfo struct {
	Id       string `json:"id"`
	IPv4 Addr `json:"IPv4"`
	// IPv4Wired map[string]string `json:"IPv4Wired"`
	// IPv4Wireless map[string]string `json:"IPv4Wireless"`
}

type DBInterface interface {
	Get(id string) (NetworkInfo, error)
	GetList() ([]NetworkInfo, error)
	Set(conf NetworkInfo) error
	Update(conf NetworkInfo) error
	Delete(id string) error
}

type Query struct {
}

var db bolt.Database

func init() {
	db = bolt.NewBoltDB(bucketName)
}

func (Query) Get(id string) (NetworkInfo, error) {
	var info NetworkInfo

	value, err := db.Get([]byte(id))
	if err != nil {
		return info, err
	}

	info, err = decode(value)
	if err != nil {
		return info, err
	}

	return info, nil
}

func (Query) GetList() ([]NetworkInfo, error) {
	infos, err := db.List()
	if err != nil {
		return nil, err
	}

	list := make([]NetworkInfo, 0)
	for _, data := range infos {
		info, err := decode([]byte(data.(string)))
		if err != nil {
			continue
		}
		list = append(list, info)
	}
	return list, nil
}

func (Query) Set(info NetworkInfo) error {
	encoded, err := info.encode()
	if err != nil {
		return err
	}

	err = db.Put([]byte(info.Id), encoded)
	if err != nil {
		return err
	}
	return nil
}

func (Query) Update(info NetworkInfo) error {
	data, err := db.Get([]byte(info.Id))
	if err != nil {
		return errors.DBOperationError{Message: err.Error()}
	}

	stored, err := decode(data)
	if err != nil {
		return err
	}

	//TODO: refactoring
	for k, v := range info.IPv4.Wired {
		stored.IPv4.Wired[k] = v
	}

	for k, v := range info.IPv4.Wireless {
		stored.IPv4.Wireless[k] = v
	}


	encoded, err := stored.encode()
	if err != nil {
		return err
	}

	return db.Put([]byte(info.Id), encoded)
}

func (Query) Delete(id string) error {
	return db.Delete([]byte(id))
}

func (info NetworkInfo) convertToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":            info.Id,
		"IPv4":      info.IPv4,
	}
}

func (info NetworkInfo) encode() ([]byte, error) {
	encoded, err := json.Marshal(info)
	if err != nil {
		return nil, errors.InvalidJSON{Message: err.Error()}
	}
	return encoded, nil
}

func decode(data []byte) (NetworkInfo, error) {
	var info NetworkInfo
	err := json.Unmarshal(data, &info)
	if err != nil {
		return info, errors.InvalidJSON{Message: err.Error()}
	}
	return info, nil
}

