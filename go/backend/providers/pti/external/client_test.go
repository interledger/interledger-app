package external

import (
	"bytes"
	"context"
	"crypto"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/pti"
)

func TestClientSignature(t *testing.T) {
	payload := `{"id":"my_internal_id-12345678","type":"Person","name":{"firstName":"Jane","lastName":"Doe"}}`

	// generated with keygen.py
	privateKeyJWK := `{"d":"CvpgDgxviQACa62u6YPOPQiA7oqwn0kLyJieTh2dokgVpyU-7ci7jS0eNp0X58Uic-LAF6TYxxlrB0bGjScfZJErX_N56f-OJOIv-7Op1vgYEgK_0z4Pa0mAbMEG_7llzZAG1wmLd1yEvwafG33HH5uoaLs9oAVjXIfI1z1PSCRdYmFoDtpx8O16hIBTtY_hDcSdI6e0bOVoiYYEHAaTum7-HhwyOVtxbijQH6rUyDsdW-Oz_Ta1XVuzVlVAedBoxRUymXwBGB6xv0rWccBopP1mAX7QUmBtHBQS9VnWILnxdPPKbx89ZLyjrt-71ow_o_JIB_QJgB3W3YBRYpBRScGWMfjPjcZWRn2XPprjCpdXzpIKstJ1lFCRSS2lyNaLLvuxnKadGrWD5DCxx79mrk5nv7A7xXVwmm4RrLKalkXsNUR6GoKvxiO50Lvdu4lXcC3WpCfZW-aoe5BCQHwbqTiq5kQUlYVOXDx9RZlY-AwNA4Bjr0tFim9zNOwFPQ2Uks46OvWUWjK4FxP0kIH9btMeZz_y4Gu2aTNgjmFHJgNluG-p-__gYXAVATjV2E7xuU6Xiv--Wvpvhz1w3leVNEYMp_ym2hPaLXB9iar9Lns4Wp_zD4XiNuI7D1lGlstzV5RlvOWVoDoAXkSZ6w2cpVL2tYzUoqlbhm3HSnNFG-k","dp":"SssUKAu5LZ0r0RjD0kzDZva_SlKWTE8cX11h65NKAswmCoz6qqJpcqJ5G7L6oPZUC8pXpajLeRPAsGuy8QsU4Rtz2WAhxVKaoUJN_o8VDUuOEGZeA-GQXOQIIIhEcUMZ38RPzo9Kg177lWh7RCUUXau7Ntwov8FCqKp2gk_UbO1AFf3p8fsIOG0MgJCYzkll0D2qSh6tPIgPaDLeJnfx0Y2QtOBvq9kIgL9pPIf1cOeIVschbbJgdRAztP_hW4htQIXnPIJW_wpXDyGSaSgvKvYAwtNBdRZHL1p26DBQ4clydWPDK1ztvUJFwYgqBrGQkmhBv4plQSUN90YoC0F57w","dq":"x4N_lLLR-_Vx_4-44rSXYD_50sW4nSBXLrEFLoooenN-OmPgLjGyjDpiyiWU5OdDgJe2AdGo0OMuD7_rJJv638o8VNGNYh9B0JSBtuCDiznJUtGwWvRdrhc906kSl8IsylT3e2Mw-2yUuW_iQ3rnqvsmkL7-cZqdlQhLI2-fjzCAOiI2f6AbVeCj8q8Mm8sQ2OevX9y8EWbPjoI2ojiQi3AC_VGJWJe9TsLoii5J8A1OoVSwZu2hWQEArK-dP97zQHwzUe8SN6KjOz-aKScoIpbkpFZan7sQvNEqENiDhfjbZm8jzfnAQBLhZRdhw4JwSzK3o4YRv93_KIymsyDTCw","e":"AQAB","kid":"0ce699d4-f5c2-415f-aa23-5dc09fe6a699","kty":"RSA","n":"2q7YWV2nGGW7gGWq96b24i2YR95SAu5XloTRDxbba6tV_-oQG_xedMT4qjKeLiU2FKLHsfNFjzBAyxKpT_cELXxWnRcFj8Ejq_1EJ4AyoHE4CE_SZtuiuyFvwd0fBwitmC-m6SdBlnyBwRQ9yojXLdBhc0oatbDlOfsgcwRl4_XY_yC3CHFa_Z7ZF7huBFNOaBYDcdN_gIggSVdufuLhHBF5AUT3j8UUguXgmfxMzgl2ClqQLabfet6YW41eY_qZP82Nq4ssDQ5OlrBM7HU3CzVM8oZJ26R0IlRF1HzE6DH4IgXdraZmtZ1eHTY2VZl1greiLXJ_ABqwHTGOno6KDPdb_lEVvSPuN0u1u4K6TNTq6yBU48-xSSWlUnmQoJyCzQoEgmpVfNt3uT4LZExLYZt5fE7WqioYaaCmdkKVBmKr2MoRtaywIN2iWQFF5pdEdxnFr7g8gzKvdmSrDfkTA7Ok3B5Y7niLEgG7IHaj2QsvxmFhfJ4YnwjGjNmhA2O4Wkmo08VAuO4PUJXXTI2vZ8SY7L16q1DzCmqohad4SVqn2wmiTEw-TZFTKNIkhBgutYh3p6-w8SsySfIFyWn2bpRUPpT8FXy4KiuFGdPEkG_OXgiA_0XJ7Vj2mzhcLPPvcRaxez8TcxMK0zsshnOkGvnD3M9uqFSUWOfNrR3tiwk","p":"8duIWDdKU8aSEpNRkv-nuoELmLL4F2k9lX07I7liqiv-NxXuSTMlPOUUTQB_LAHU5zwbB_nZpkg_nSBkM4_BtFefjqvNFrQMBflLwuC1CZ7Bb63tNSj96-vNe0y_dduI8pm36UCMFxvqLYyvIRX_ZzZQPe-zWgyWCNSzraxJ4Lh7ONqSjD7umYmn4PDSX7-Uzj0Knsndzzvo6cJ24MoBJ96UpsyPMVwTEhMMQ6sxT6cxIJhJUqt8ReKUIXFSDwDbMQl1NjIj6nFHhle02kCZ1tyds1Pp7LifoW_nlEUU9O0rcym6fbvhW9FNUGxFQYn9Cz1gt0H702pIIRiCRcUGyw","q":"53hnHBOgQ3e6FkujgbsnBMBALvv1lvJkaZfZK0fgFJMSI_cccA3aS2xiSg-vWB9mBmeL6z4aV3Sq-Ra9iJ4tdBao6oXj2w8opT3ygFfbQSkXw2VCtRGsLk02oGVAen2P5Kv7q017mZ-XCIJREm-01Yki7TIBXw97TkBA4HddUbIvdeFN18vOctifNci3m4AF9rrwuJeQJTXkhAew8sJBxn5HDnRYRaaFyFvd2lXAl4FtqQMxmXQMaT2HDWuN4rKrQZrt_twYREo1Yn6eDB_IgHJO3CA1LVuht_9XpEAsOVtdKHxKsKZURgpZLX-yrnfig5wTjTBK7e4q0cdbgsFm-w","qi":"GwiMb1ZFIR98_HjIBqAd0tQlNTHq2dFru0ZM8RwFB4aSanIXUt5WxjUIN85NNPY5MMiILqk0ZhyPMBIsYEzWQFenu9irL-j0Dp2QDJ6Yi1xlME1xW4VQyZJehGlpNXHVLiTglweJnkx4NKMgwk5npflNvn3jGOn3pYtSAhKLiSKUkDexXJba7I7140dX7QFpH7wp-a3kGFVB1SkgSSz1nU4dOVFof0KMvrMqHaMfTkglTwalmaW8dsZbs8AXoCRYXpOmJDUvMgbqeHiCAtbBIhk-8Gs23AiPJxnoHjaXi5nwGm8CoeCnXVEHPNM26iLiFIn5sb3RiL_giEnV0-wUXQ"}`

	// generated with signed_request_maker.py
	expectedSignature := "eyJhbGciOiJSUzUxMiIsImNpZCI6InRlc3QxMjMiLCJraWQiOiJhemlDcTQyM1NMeXVsc1oxdGtUZ2tuaEJ2M2VveGkxbGc1Y3BxbnVldVlBIn0.UE9TVAowQkEzMUNEODM2MERGRTA5RUExN0U5OEVGODZBRDlEQTBGNkUyQTA4MTVFNjVDOEFGNEUwN0Y4QTU3NThERDVGCmNvbnRlbnQtdHlwZTphcHBsaWNhdGlvbi9qc29uCmRhdGU6VHVlLCAwOSBKYW4gMjAyNCAxMjoyMToxMSBHTVQKeC1wdGktY2xpZW50LWlkOnRlc3QxMjMKL3YxL3VzZXJz.TkUAGZW4DhqIIcHC_eITpYlvIXpxfavFOqoXPuk4FznGzX_S-yVp4Y-uffWKOwuyt3EF0ES1eBpLPW0qQ3ntqvXugwZ7hXnqL0CYMoq0vR_tevn7Q2zKUFvp6EkHt90wuHBckjRqJ4RJ5sW1OnRzSRQ6W4NkTCWwS4W24lTB41OG68_1vpxUJMGa0mgQMVDMi5SqS7Rv3f4pVKWx2fIoS5vvRCI4Lz4IVJ7dodKJ52EIHAa-DTSkxnG8m75dMlRIhLUkraMqAg5ahodkl8rVX5XXGi1qmLMFnqGAwFrE-SNI4toZDERRVBN9bz7jQ3SyZmzJbWjT6Lt4I0cuEZJz3A2dV1KBqSlHByCStXGAx4pG2JXd5mNL3-rGXaWwR1xX2VmIW83B6l-cKYQQBl2jrzX5_zSQi3jYfVvEYOGvB6g0VeXr3DLqrJ-O7JQEokJ5S62AIdNANucCvhjshwvksd6k-G6gmUpn9XGBFmG3e1vZ3MDEKTuTKltCU2UIhs1XHG_odRd2_KLw_qY_ygRewplpGVbDA75JXXWqhr5NEYFz0Y8RuWigFFT-FGn5RdU7G5e4O-R1KNG7DmA0o0zzYakXN5kLgR7t9KTf5Zvkt1wjR77WzHyv681IcrCgpvbqrpNhsvsFzY4fX9ya2mab2WmdjiO0mBk4B9lHox_BJYE"
	date, err := time.Parse(http.TimeFormat, "Tue, 09 Jan 2024 12:21:11 GMT")
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "https://pti.apistaging.pticlient.com/v1/users", bytes.NewBuffer([]byte(payload)))
	require.NoError(t, err)
	req.Header.Add(ptiClientIDHeader, "test123")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Date", date.Format(http.TimeFormat))

	key, err := jwk.ParseKey([]byte(privateKeyJWK))
	require.NoError(t, err)

	// we remove the `kid` from the jwk otherwise lestrat-jws won't let us override it in the protected header
	require.NoError(t, key.Remove("kid"))
	pubKey, err := key.PublicKey()
	require.NoError(t, err)
	thumb, err := pubKey.Thumbprint(crypto.SHA256)
	require.NoError(t, err)

	err = Sign(context.Background(), req, date, []byte(payload), key, base64.RawURLEncoding.EncodeToString(thumb))
	require.NoError(t, err)

	assert.Equal(t, expectedSignature, req.Header.Get(ptiSignatureHeader))
}

func TestVerify(t *testing.T) {
	payload := `{"id":"my_internal_id-12345678","type":"Person","name":{"firstName":"Jane","lastName":"Doe"}}`
	privateKeyJWK := `{"d":"CvpgDgxviQACa62u6YPOPQiA7oqwn0kLyJieTh2dokgVpyU-7ci7jS0eNp0X58Uic-LAF6TYxxlrB0bGjScfZJErX_N56f-OJOIv-7Op1vgYEgK_0z4Pa0mAbMEG_7llzZAG1wmLd1yEvwafG33HH5uoaLs9oAVjXIfI1z1PSCRdYmFoDtpx8O16hIBTtY_hDcSdI6e0bOVoiYYEHAaTum7-HhwyOVtxbijQH6rUyDsdW-Oz_Ta1XVuzVlVAedBoxRUymXwBGB6xv0rWccBopP1mAX7QUmBtHBQS9VnWILnxdPPKbx89ZLyjrt-71ow_o_JIB_QJgB3W3YBRYpBRScGWMfjPjcZWRn2XPprjCpdXzpIKstJ1lFCRSS2lyNaLLvuxnKadGrWD5DCxx79mrk5nv7A7xXVwmm4RrLKalkXsNUR6GoKvxiO50Lvdu4lXcC3WpCfZW-aoe5BCQHwbqTiq5kQUlYVOXDx9RZlY-AwNA4Bjr0tFim9zNOwFPQ2Uks46OvWUWjK4FxP0kIH9btMeZz_y4Gu2aTNgjmFHJgNluG-p-__gYXAVATjV2E7xuU6Xiv--Wvpvhz1w3leVNEYMp_ym2hPaLXB9iar9Lns4Wp_zD4XiNuI7D1lGlstzV5RlvOWVoDoAXkSZ6w2cpVL2tYzUoqlbhm3HSnNFG-k","dp":"SssUKAu5LZ0r0RjD0kzDZva_SlKWTE8cX11h65NKAswmCoz6qqJpcqJ5G7L6oPZUC8pXpajLeRPAsGuy8QsU4Rtz2WAhxVKaoUJN_o8VDUuOEGZeA-GQXOQIIIhEcUMZ38RPzo9Kg177lWh7RCUUXau7Ntwov8FCqKp2gk_UbO1AFf3p8fsIOG0MgJCYzkll0D2qSh6tPIgPaDLeJnfx0Y2QtOBvq9kIgL9pPIf1cOeIVschbbJgdRAztP_hW4htQIXnPIJW_wpXDyGSaSgvKvYAwtNBdRZHL1p26DBQ4clydWPDK1ztvUJFwYgqBrGQkmhBv4plQSUN90YoC0F57w","dq":"x4N_lLLR-_Vx_4-44rSXYD_50sW4nSBXLrEFLoooenN-OmPgLjGyjDpiyiWU5OdDgJe2AdGo0OMuD7_rJJv638o8VNGNYh9B0JSBtuCDiznJUtGwWvRdrhc906kSl8IsylT3e2Mw-2yUuW_iQ3rnqvsmkL7-cZqdlQhLI2-fjzCAOiI2f6AbVeCj8q8Mm8sQ2OevX9y8EWbPjoI2ojiQi3AC_VGJWJe9TsLoii5J8A1OoVSwZu2hWQEArK-dP97zQHwzUe8SN6KjOz-aKScoIpbkpFZan7sQvNEqENiDhfjbZm8jzfnAQBLhZRdhw4JwSzK3o4YRv93_KIymsyDTCw","e":"AQAB","kid":"0ce699d4-f5c2-415f-aa23-5dc09fe6a699","kty":"RSA","n":"2q7YWV2nGGW7gGWq96b24i2YR95SAu5XloTRDxbba6tV_-oQG_xedMT4qjKeLiU2FKLHsfNFjzBAyxKpT_cELXxWnRcFj8Ejq_1EJ4AyoHE4CE_SZtuiuyFvwd0fBwitmC-m6SdBlnyBwRQ9yojXLdBhc0oatbDlOfsgcwRl4_XY_yC3CHFa_Z7ZF7huBFNOaBYDcdN_gIggSVdufuLhHBF5AUT3j8UUguXgmfxMzgl2ClqQLabfet6YW41eY_qZP82Nq4ssDQ5OlrBM7HU3CzVM8oZJ26R0IlRF1HzE6DH4IgXdraZmtZ1eHTY2VZl1greiLXJ_ABqwHTGOno6KDPdb_lEVvSPuN0u1u4K6TNTq6yBU48-xSSWlUnmQoJyCzQoEgmpVfNt3uT4LZExLYZt5fE7WqioYaaCmdkKVBmKr2MoRtaywIN2iWQFF5pdEdxnFr7g8gzKvdmSrDfkTA7Ok3B5Y7niLEgG7IHaj2QsvxmFhfJ4YnwjGjNmhA2O4Wkmo08VAuO4PUJXXTI2vZ8SY7L16q1DzCmqohad4SVqn2wmiTEw-TZFTKNIkhBgutYh3p6-w8SsySfIFyWn2bpRUPpT8FXy4KiuFGdPEkG_OXgiA_0XJ7Vj2mzhcLPPvcRaxez8TcxMK0zsshnOkGvnD3M9uqFSUWOfNrR3tiwk","p":"8duIWDdKU8aSEpNRkv-nuoELmLL4F2k9lX07I7liqiv-NxXuSTMlPOUUTQB_LAHU5zwbB_nZpkg_nSBkM4_BtFefjqvNFrQMBflLwuC1CZ7Bb63tNSj96-vNe0y_dduI8pm36UCMFxvqLYyvIRX_ZzZQPe-zWgyWCNSzraxJ4Lh7ONqSjD7umYmn4PDSX7-Uzj0Knsndzzvo6cJ24MoBJ96UpsyPMVwTEhMMQ6sxT6cxIJhJUqt8ReKUIXFSDwDbMQl1NjIj6nFHhle02kCZ1tyds1Pp7LifoW_nlEUU9O0rcym6fbvhW9FNUGxFQYn9Cz1gt0H702pIIRiCRcUGyw","q":"53hnHBOgQ3e6FkujgbsnBMBALvv1lvJkaZfZK0fgFJMSI_cccA3aS2xiSg-vWB9mBmeL6z4aV3Sq-Ra9iJ4tdBao6oXj2w8opT3ygFfbQSkXw2VCtRGsLk02oGVAen2P5Kv7q017mZ-XCIJREm-01Yki7TIBXw97TkBA4HddUbIvdeFN18vOctifNci3m4AF9rrwuJeQJTXkhAew8sJBxn5HDnRYRaaFyFvd2lXAl4FtqQMxmXQMaT2HDWuN4rKrQZrt_twYREo1Yn6eDB_IgHJO3CA1LVuht_9XpEAsOVtdKHxKsKZURgpZLX-yrnfig5wTjTBK7e4q0cdbgsFm-w","qi":"GwiMb1ZFIR98_HjIBqAd0tQlNTHq2dFru0ZM8RwFB4aSanIXUt5WxjUIN85NNPY5MMiILqk0ZhyPMBIsYEzWQFenu9irL-j0Dp2QDJ6Yi1xlME1xW4VQyZJehGlpNXHVLiTglweJnkx4NKMgwk5npflNvn3jGOn3pYtSAhKLiSKUkDexXJba7I7140dX7QFpH7wp-a3kGFVB1SkgSSz1nU4dOVFof0KMvrMqHaMfTkglTwalmaW8dsZbs8AXoCRYXpOmJDUvMgbqeHiCAtbBIhk-8Gs23AiPJxnoHjaXi5nwGm8CoeCnXVEHPNM26iLiFIn5sb3RiL_giEnV0-wUXQ"}`
	signature := "eyJhbGciOiJSUzUxMiIsImNpZCI6InRlc3QxMjMiLCJraWQiOiJhemlDcTQyM1NMeXVsc1oxdGtUZ2tuaEJ2M2VveGkxbGc1Y3BxbnVldVlBIn0.UE9TVAowQkEzMUNEODM2MERGRTA5RUExN0U5OEVGODZBRDlEQTBGNkUyQTA4MTVFNjVDOEFGNEUwN0Y4QTU3NThERDVGCmNvbnRlbnQtdHlwZTphcHBsaWNhdGlvbi9qc29uCmRhdGU6VHVlLCAwOSBKYW4gMjAyNCAxMjoyMToxMSBHTVQKeC1wdGktY2xpZW50LWlkOnRlc3QxMjMKL3YxL3VzZXJz.TkUAGZW4DhqIIcHC_eITpYlvIXpxfavFOqoXPuk4FznGzX_S-yVp4Y-uffWKOwuyt3EF0ES1eBpLPW0qQ3ntqvXugwZ7hXnqL0CYMoq0vR_tevn7Q2zKUFvp6EkHt90wuHBckjRqJ4RJ5sW1OnRzSRQ6W4NkTCWwS4W24lTB41OG68_1vpxUJMGa0mgQMVDMi5SqS7Rv3f4pVKWx2fIoS5vvRCI4Lz4IVJ7dodKJ52EIHAa-DTSkxnG8m75dMlRIhLUkraMqAg5ahodkl8rVX5XXGi1qmLMFnqGAwFrE-SNI4toZDERRVBN9bz7jQ3SyZmzJbWjT6Lt4I0cuEZJz3A2dV1KBqSlHByCStXGAx4pG2JXd5mNL3-rGXaWwR1xX2VmIW83B6l-cKYQQBl2jrzX5_zSQi3jYfVvEYOGvB6g0VeXr3DLqrJ-O7JQEokJ5S62AIdNANucCvhjshwvksd6k-G6gmUpn9XGBFmG3e1vZ3MDEKTuTKltCU2UIhs1XHG_odRd2_KLw_qY_ygRewplpGVbDA75JXXWqhr5NEYFz0Y8RuWigFFT-FGn5RdU7G5e4O-R1KNG7DmA0o0zzYakXN5kLgR7t9KTf5Zvkt1wjR77WzHyv681IcrCgpvbqrpNhsvsFzY4fX9ya2mab2WmdjiO0mBk4B9lHox_BJYE"
	key, err := jwk.ParseKey([]byte(privateKeyJWK))
	require.NoError(t, err)
	publicKey, err := key.PublicKey()
	require.NoError(t, err)
	date, err := time.Parse(http.TimeFormat, "Tue, 09 Jan 2024 12:21:11 GMT")
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "https://pti.apistaging.pticlient.com/v1/users", bytes.NewBuffer([]byte(payload)))
	require.NoError(t, err)
	req.Header.Add(ptiClientIDHeader, "test123")
	req.Header.Add(ptiSignatureHeader, signature)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Date", date.Format(http.TimeFormat))

	err = Verify(context.Background(), req, publicKey)
	require.NoError(t, err)
}

func TestPutUser(t *testing.T) {
	if os.Getenv("PTI_JWK") == "" || os.Getenv("PTI_CLIENTI_ID") == "" {
		t.Skip("no credentials")
	}
	key, err := jwk.ParseKey([]byte(os.Getenv("PTI_JWK")))
	require.NoError(t, err)

	client := New(ClientArgs{
		PrivateKey: key,
		ClientID:   os.Getenv("PTI_CLIENT_ID"),
	})

	u, err := client.PutUser(context.Background(), PutUserArgs{
		ID:   "031823ff-3db5-40b5-9c4a-7bede87edfc0",
		Type: "PERSON",
		Phones: []Phone{
			{
				Number:  "+270823855973",
				Type:    "MOBILE",
				Default: true,
			},
		},
		DateOfBirth: "1981-01-01",
		Name: Name{
			First: "Jimmy",
			Last:  "Prawns",
		},
		Emails: []Email{{
			Address: "jimmy@openpayments.dev",
			Default: true,
		}},
		Addresses: []Address{
			{
				Street:     "785 Market Street",
				City:       "San Francisco",
				PostalCode: "94103",
				StateCode: StateCode{
					Code: "US-CA",
				},
				Country: CountryCode{
					Code: "US",
				},
				Default: true,
			},
		},
	})
	require.NoError(t, err)

	usr, err := client.GetUser(context.Background(), u)
	require.NoError(t, err)
	fmt.Printf("updated user: %+v", usr)
}

func TestStartAssessment(t *testing.T) {
	if os.Getenv("PTI_JWK") == "" || os.Getenv("PTI_CLIENTI_ID") == "" {
		t.Skip("no credentials")
	}
	key, err := jwk.ParseKey([]byte(os.Getenv("PTI_JWK")))
	require.NoError(t, err)

	client := New(ClientArgs{
		PrivateKey: key,
		ClientID:   os.Getenv("PTI_CLIENT_ID"),
	})

	a, err := client.StartUserAssessment(context.Background(), StartUserAssessmentArgs{
		ID:         "031823ff-3db5-40b5-9c4a-7bede87edfc0",
		Type:       "PERSON",
		ScenarioID: pti.ScenarioDeposit,
	})
	require.NoError(t, err)

	fmt.Println("assessment: ", a)
}

func TestGetAssessment(t *testing.T) {
	if os.Getenv("PTI_JWK") == "" || os.Getenv("PTI_CLIENTI_ID") == "" {
		t.Skip("no credentials")
	}
	key, err := jwk.ParseKey([]byte(os.Getenv("PTI_JWK")))
	require.NoError(t, err)

	client := New(ClientArgs{
		PrivateKey: key,
		ClientID:   os.Getenv("PTI_CLIENT_ID"),
	})

	a, err := client.GetUserAssessment(context.Background(), "031823ff-3db5-40b5-9c4a-7bede87edfc0")
	require.NoError(t, err)

	fmt.Printf("assessment: %+v\n", a)
}
