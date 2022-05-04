package fctmultipart

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
)

const fileJSON = `[
          ["key", "tmp/59560132796/bulk/0260173a-6418-46ca-9198-3328a8001e81/7b32a8a8-3c87-46ab-b570-9c1a4d09c0a0"],
          ["Content-Type", "application/x-jsonlines"],
          ["success_action_status", "201"],
          ["acl", "private"],
          ["policy", "eyJleHBpcmF0aW9uIjoiMjAyMS0xMS0xOFQxNjoyNjoxMVoiLCJjb25kaXRpb25zIjpbeyJidWNrZXQiOiJzaG9waWZ5In0sWyJjb250ZW50LWxlbmd0aC1yYW5nZSIsMSwyMDk3MTUyMF0seyJrZXkiOiJ0bXAvNTk1NjAxMzI3OTYvYnVsay8wMjYwMTczYS02NDE4LTQ2Y2EtOTE5OC0zMzI4YTgwMDFlODEvN2IzMmE4YTgtM2M4Ny00NmFiLWI1NzAtOWMxYTRkMDljMGEwIn0seyJDb250ZW50LVR5cGUiOiJhcHBsaWNhdGlvbi94LWpzb25saW5lcyJ9LHsic3VjY2Vzc19hY3Rpb25fc3RhdHVzIjoiMjAxIn0seyJhY2wiOiJwcml2YXRlIn0seyJ4LWFtei1jcmVkZW50aWFsIjoiQUtJQUpZTTU1NUtWWUVXR0pES1EvMjAyMTExMTgvdXMtZWFzdC0xL3MzL2F3czRfcmVxdWVzdCJ9LHsieC1hbXotYWxnb3JpdGhtIjoiQVdTNC1ITUFDLVNIQTI1NiJ9LHsieC1hbXotZGF0ZSI6IjIwMjExMTE4VDE1MjYxMVoifV19"],
          ["x-amz-credential", "AKIAJYM555KVYEWGJDKQ/20211118/us-east-1/s3/aws4_request"],
          ["x-amz-algorithm", "AWS4-HMAC-SHA256"],
          ["x-amz-date", "20211118T152611Z"],
          ["x-amz-signature", "d76b0e25c07d6bf6272b8ba73fc66591bf8ae91a6d435cc909e2d697fc1a8806"]
]`

const UpdateFile = `{"input":{"id":"gid://shopify/ProductVariant/41078819979452","price":"26.99","compareAtPrice":"29.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41078852649148","price":"11.49","compareAtPrice":"12.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41078866084028","price":"35.99","compareAtPrice":"35.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41078907044028","price":"13.99","compareAtPrice":"15.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41078907371708","price":"15.99","compareAtPrice":"17.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41078913106108","price":"11.99","compareAtPrice":"10.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41078946365628","price":"12.49","compareAtPrice":"12.49"}}
{"input":{"id":"gid://shopify/ProductVariant/41078951575740","price":"13.49","compareAtPrice":"13.49"}}
{"input":{"id":"gid://shopify/ProductVariant/41078955966652","price":"13.49","compareAtPrice":"13.49"}}
{"input":{"id":"gid://shopify/ProductVariant/41078975955132","price":"25.99","compareAtPrice":"25.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41078979461308","price":"14.64","compareAtPrice":"14.64"}}
{"input":{"id":"gid://shopify/ProductVariant/41078982836412","price":"11.99","compareAtPrice":"11.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41079002267836","price":"4.49","compareAtPrice":"4.49"}}
{"input":{"id":"gid://shopify/ProductVariant/41079002759356","price":"15.99","compareAtPrice":"15.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41079022682300","price":"20.69","compareAtPrice":"22.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41079031005372","price":"22.49","compareAtPrice":"22.49"}}
{"input":{"id":"gid://shopify/ProductVariant/41079032119484","price":"11.69","compareAtPrice":"11.69"}}
{"input":{"id":"gid://shopify/ProductVariant/41079035789500","price":"13.49","compareAtPrice":"13.49"}}
{"input":{"id":"gid://shopify/ProductVariant/41079050830012","price":"35.79","compareAtPrice":"35.79"}}
{"input":{"id":"gid://shopify/ProductVariant/41079051649212","price":"35.79","compareAtPrice":"35.79"}}
{"input":{"id":"gid://shopify/ProductVariant/41079052828860","price":"35.79","compareAtPrice":"35.79"}}
{"input":{"id":"gid://shopify/ProductVariant/41079063118012","price":"16.04","compareAtPrice":"16.04"}}
{"input":{"id":"gid://shopify/ProductVariant/41079084286140","price":"9.49","compareAtPrice":"9.49"}}
{"input":{"id":"gid://shopify/ProductVariant/41079086252220","price":"18.92","compareAtPrice":"18.92"}}
{"input":{"id":"gid://shopify/ProductVariant/41079091855548","price":"17.49","compareAtPrice":"17.49"}}
{"input":{"id":"gid://shopify/ProductVariant/41079098572988","price":"11.49","compareAtPrice":"11.49"}}
{"input":{"id":"gid://shopify/ProductVariant/41079098638524","price":"8.99","compareAtPrice":"8.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41079098704060","price":"8.99","compareAtPrice":"8.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41079100866748","price":"13.49","compareAtPrice":"13.49"}}
{"input":{"id":"gid://shopify/ProductVariant/41079103783100","price":"15.99","compareAtPrice":"15.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41079106109628","price":"26.99","compareAtPrice":"26.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41079113875644","price":"16.99","compareAtPrice":"16.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41079118856380","price":"8.99","compareAtPrice":"8.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41079126261948","price":"14.24","compareAtPrice":"14.24"}}
{"input":{"id":"gid://shopify/ProductVariant/41079132487868","price":"19.74","compareAtPrice":"69.69"}}
{"input":{"id":"gid://shopify/ProductVariant/41079172563132","price":"13.99","compareAtPrice":"13.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41079187177660","price":"27.49","compareAtPrice":"27.49"}}
{"input":{"id":"gid://shopify/ProductVariant/41079187275964","price":"35.79","compareAtPrice":"39.79"}}
{"input":{"id":"gid://shopify/ProductVariant/41079187439804","price":"35.99","compareAtPrice":"35.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41079196156092","price":"11.69","compareAtPrice":"12.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41079198318780","price":"11.69","compareAtPrice":"11.69"}}
{"input":{"id":"gid://shopify/ProductVariant/41079200317628","price":"15.09","compareAtPrice":"16.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41079206838460","price":"17.99","compareAtPrice":"19.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41079211065532","price":"14.09","compareAtPrice":"14.09"}}
{"input":{"id":"gid://shopify/ProductVariant/41079213457596","price":"14.39","compareAtPrice":"14.39"}}
{"input":{"id":"gid://shopify/ProductVariant/41079218241724","price":"53.99","compareAtPrice":"53.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41079220502716","price":"35.79","compareAtPrice":"35.79"}}
{"input":{"id":"gid://shopify/ProductVariant/41079256842428","price":"35.79","compareAtPrice":"35.79"}}
{"input":{"id":"gid://shopify/ProductVariant/41079261954236","price":"12.49","compareAtPrice":"12.49"}}
{"input":{"id":"gid://shopify/ProductVariant/41079289020604","price":"8.79","compareAtPrice":"8.79"}}
{"input":{"id":"gid://shopify/ProductVariant/41079292723388","price":"14.24","compareAtPrice":"14.24"}}
{"input":{"id":"gid://shopify/ProductVariant/41079295606972","price":"35.99","compareAtPrice":"36.89"}}
{"input":{"id":"gid://shopify/ProductVariant/41079301734588","price":"9.99","compareAtPrice":"9.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41079306453180","price":"8.99","compareAtPrice":"8.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41079310909628","price":"21.49","compareAtPrice":"10.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41079336304828","price":"8.99","compareAtPrice":"8.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41079339548860","price":"8.99","compareAtPrice":"8.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41079341580476","price":"17.99","compareAtPrice":"17.99"}}
{"input":{"id":"gid://shopify/ProductVariant/41079347740860","price":"21.59","compareAtPrice":"21.59"}}`

func TestMultipartForm(t *testing.T) {
	var file [][]string

	err := json.Unmarshal([]byte(fileJSON), &file)
	if err != nil {
		t.Error(err)
		return
	}

	fileBody := bytes.NewReader([]byte(UpdateFile))

	form, err := BuildMultipartForm(file, "test_file", fileBody)
	if err != nil {
		t.Error(err)
		return
	}

	formBody, err := ioutil.ReadAll(form.GetReader())
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(string(formBody))
}
