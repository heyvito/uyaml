/*
Package uyaml provides utilities for working with unstructured yaml documents.
It allows structures to be queried, modified, and removed using a small query
language that somewhat resembles CSS selectors (in some parallel dimension).
This package is considered of beta quality, and may behave oddly in some edge
cases. In case this happens, please file an issue to the corresponding
repository.

Querying Values

Let's suppose you have a extensive YAML document with a key nested under some
complicated hierarchy. One option is to implement structures matching the
document's, so yaml.Unmarshal can do its job. The other option is to unmarshal
into a map of interfaces, or slice of interfaces. uyaml implements two methods
to handle this scenario: DigItem and MustDigItem. For example, consider the
following YAML document:

	userCount: 2
	users:
	- name: josie
	  roles:
	  - bot
	  - foo
	  - bar
	- name: lester
	  roles:
	  - dummy

In order to obtain roles for a user under the key 'josie', DigItem can be used:

	doc, err := uyaml.Decode(...)
	if err != nil {
		...
	}
	ok, item, err := doc.DigItem("users.(name='josie').roles")
	if ok {
		fmt.Printf("%#v", item.([]interface{}))
	}

MustDigItem works just like DigItem, except it only have a single return value,
and panics in case the item cannot be found or the provided path cannot be
parsed.

Removing Values

Removing values can be done with the Remove method, which takes a single path
to be removed. In case the path does not exist, a noop happens and the same
object is returned.

	doc, err := uyaml.Decode(...)
	if err != nil {
		...
	}
	doc, err := doc.Remove("users.(name='josie')")

Setting Values

Set can be used to inject arbitrary values into the document's structure. For
instance (errors checking omitted for brevity):

	val, err := data.Set("users.(name='dummy').test", true)
	yam, err := val.Encode()
	fmt.Println(yam)

Would print the following structure:

	userCount: 2
	users:
	- name: josie
	  roles:
	  - bot
	  - foo
	  - bar
	- name: lester
	  roles:
	  - dummy
	- name: dummy
	  test: true
*/
package uyaml
