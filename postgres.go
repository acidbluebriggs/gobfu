/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gobfuscator

import "fmt"

var Postgres = SQLFormatter{

	PhoneNumber: func(item *Item) string {
		return fmt.Sprintf("%s = lpad(\"%s\".id::text || \"%s\".id::text, 15, '0')",
			item.column, item.table, item.table)
	},
	Email: func(item *Item) string {
		return fmt.Sprintf("%s = \"%s\".id::text || '@example.com'", item.column, item.table)
	},
	Word: func(item *Item) string {
		return fmt.Sprintf("%s = ''", item.column)
	},
	Null: func(item *Item) string {
		return fmt.Sprintf("%s = NULL", item.column)
	},
	Address: func(item *Item) string {
		return fmt.Sprintf("%s = \"%s\".id::text || ' ' || (select value from gobfuscator_anon_data "+
			"where kind = '%s' "+
			"and idx = mod(\"%s\".id, (select total from gobfuscater.gobfuscater_data_census where kind = '%s')) "+
			"limit 1)%s",
			item.column, item.table, item.generator, item.table, item.generator, cast(item.sqlType))
	},
	Business: func(item *Item) string {
		// We'll use the global cast function here since we can't access p.cast during initialization
		return fmt.Sprintf("%s = (select value "+
			"from gobfuscator_anon_data "+
			"where kind = '%s' "+
			"and idx = mod(\"%s\".id, (select total from gobfuscater.gobfuscater_data_census where kind = '%s')) "+
			"limit 1)%s || ' ' || \"%s\".id::text",
			item.column, item.generator, item.table, item.generator, cast(item.sqlType), item.table)
	},
	Default: func(item *Item) string {
		return fmt.Sprintf("%s = (select value "+
			"from gobfuscator_anon_data "+
			"where kind = '%s' "+
			"and idx = mod(\"%s\".id, (select total from gobfuscater.gobfuscater_data_census where kind = '%s')) "+
			"limit 1)%s || ' ' || \"%s\".id::text",
			item.column, item.generator, item.table, item.generator, cast(item.sqlType), item.table)
	},
}

func cast(sqlType string) string {
	switch sqlType {
	case "jsonb":
		return "::jsonb"
	case "date":
		return "::date"
	default:
		return ""
	}
}
