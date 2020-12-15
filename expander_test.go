package swagexpander_test

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"testing"

	"gopkg.in/yaml.v2"
	"github.com/breise/swagexpander"
)

type testCase struct {
	name     string
	inp      string
	exp      string
	inpFile  string
	expFile  string
	expError string
}

func TestExpander(t *testing.T) {

	var tests = []testCase{
		{
			name: "string", inp: `
type: string
`,
			exp: `
type: string
`,
		},
		{
			name: "simple object", inp: `
type: object
properties:
  thing:
    type: string
`,
			exp: `
type: object
properties:
  thing:
    type: string
`,
		},
		{
			name: "array of numbers",
			inp: `
type: array
items:
  type: number
`,
			exp: `
type: array
items:
  type: number
`,
		},
		{
			name: "array of objects",
			inp: `
type: array
items:
  type: object
  properties:
    id:
      type: integer
    name:
      type: string
`,
			exp: `
type: array
items:
  type: object
  properties:
    id:
      type: integer
    name:
      type: string
`,
		},
		{
			name:    "petstore post pet",
			inpFile: "data/petstore_pet_inp.yaml",
			expFile: "data/petstore_pet_exp.yaml",
		},
		{
			name:     "petstore post pet with one-level cycle",
			inpFile:  "data/petstore_pet_with_cycle_inp.yaml",
			expError: "cannot copyAndExpand(): $ref cycle detected: #/definitions/Pet -> #/definitions/Category -> #/definitions/Pet",
		},
		{
			name:     "petstore post pet with two-level cycle",
			inpFile:  "data/petstore_pet_with_two_level_cycle_inp.yaml",
			expError: "cannot copyAndExpand(): $ref cycle detected: #/definitions/Pet -> #/definitions/Tag -> #/definitions/Category -> #/definitions/Pet",
		},
	}
	for i, tc := range tests {
		desc := fmt.Sprintf("Test Case %d: %s", i, tc.name)
		expError := tc.expError
		inp, err := readFileOrString(tc.inpFile, tc.inp)
		if err != nil {
			t.Fatal(err)
		}
		t.Run(desc, func(t *testing.T) {
			var thing map[interface{}]interface{}
			if err := yaml.Unmarshal(inp, &thing); err != nil {
				t.Fatalf("cannot unmarshal '%s'. Error: %s", inp, err)
			}
			got, gotErr := swagexpander.CopyAndExpand(thing)
			if expError == "" {
				if gotErr != nil {
					t.Fatal(gotErr)
				}
				exp, err := readFileOrString(tc.expFile, tc.exp)
				if err != nil {
					t.Fatal(err)
				}
				var expObj map[interface{}]interface{}
				if err := yaml.Unmarshal(exp, &expObj); err != nil {
					t.Fatalf("cannot unmarshal expected response: %s", err)
				}
				if !reflect.DeepEqual(got, expObj) {
					gotYaml, err := yaml.Marshal(got)
					if err != nil {
						t.Fatalf("cannot yaml.Marshal(got): %s", err)
					}
					expYaml, err := yaml.Marshal(expObj)
					if err != nil {
						t.Fatalf("cannot yaml.Marshal(expObj): %s", err)
					}
					t.Errorf("no match:\nGot: %s\nExp: %s", gotYaml, expYaml)
					gotDisp := fmt.Sprintf("%+v", got)
					expDisp := fmt.Sprintf("%+v", expObj)
					t.Errorf("no match:\nGot:\nType: %T\nValue: %s\nExp:\nType: %T\nValue: %s", got, gotDisp, expObj, expDisp)
				}
			} else {
				// expecting error
				if gotErr == nil {
					t.Fatalf("%s: Expecting error but SUCCEEDED", desc)
				}
				expErrorRe := regexp.MustCompile(regexp.QuoteMeta(expError))
				if !expErrorRe.MatchString(gotErr.Error()) {
					t.Errorf("%s: FAILS OK, but error message did not match.\nGot: %s\nExp: %s", desc, gotErr, expError)
				}
			}
		})
	}
}

func readFileOrString(file, s string) ([]byte, error) {
	var rv []byte
	if file != "" {
		var err error
		rv, err = ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("cannot open file '%s' for reading: %s", file, err)
		}
	} else {
		rv = []byte(s)
	}
	return rv, nil
}
