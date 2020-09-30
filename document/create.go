/*
 * Copyright 2020 Luke Whrit, Jack Dorland; The Spacebin Authors

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 *     http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package document

import (
	b64 "encoding/base64"

	"github.com/gofiber/fiber/v2"
)

func registerCreate(api fiber.Router) {
	api.Post("/", func(c *fiber.Ctx) error {
		b := new(CreateRequest)

		// Validate and parse body
		if err := c.BodyParser(b); err != nil {
			c.Status(400).JSON(&Response{
				Status:  c.Response().StatusCode(),
				Payload: Payload{},
				Error:   err.Error(),
			})

			return nil
		}

		if err := b.Validate(); err != nil {
			c.Status(400).JSON(&Response{
				Status:  c.Response().StatusCode(),
				Payload: Payload{},
				Error:   err.Error(),
			})

			return nil
		}

		// Create and retrieve document
		id, err := NewDocument(b.Content, b.Extension)

		if err != nil {
			c.Status(500).JSON(&Response{
				Status:  c.Response().StatusCode(),
				Payload: Payload{},
				Error:   err.Error(),
			})

			return nil
		}

		document, err := GetDocument(id)

		if err != nil {
			c.Status(500).JSON(&Response{
				Status:  c.Response().StatusCode(),
				Payload: Payload{},
				Error:   err.Error(),
			})

			return nil
		}

		c.Status(201).JSON(&Response{
			Status: c.Response().StatusCode(),
			Payload: Payload{
				ID:          &document.ID,
				ContentHash: b64.StdEncoding.EncodeToString([]byte(document.Content)),
			},
			Error: "",
		})

		return nil
	})
}
