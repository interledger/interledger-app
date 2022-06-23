package utils

import (
	"bytes"
	b64 "encoding/base64"
	"text/template"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func StringArrayOutputFromStack(stack *pulumi.StackReference, key string) pulumi.StringArrayOutput {
	return stack.GetOutput(pulumi.String(key)).ApplyT(func(arg interface{}) []string {
		stringArray := arg.([]interface{})
		var outputs []string
		for _, s := range stringArray {
			outputs = append(outputs, s.(string))
		}
		return outputs
	}).(pulumi.StringArrayOutput)
}

func ParseTemplate(data interface{}, filePath string) string {
	tmp, err := template.ParseFiles(filePath)
	if err != nil {
		return ""
	}
	document := &bytes.Buffer{}
	err = tmp.Execute(document, data)
	if err != nil {
		panic(err)
		return ""
	}

	return b64.StdEncoding.EncodeToString(document.Bytes())
}

func ParseTemplateAsBytes(data interface{}, filePath string) (*bytes.Buffer, error) {
	tmp, err := template.ParseFiles(filePath)
	if err != nil {
		return nil, err
	}

	document := &bytes.Buffer{}
	err = tmp.Execute(document, data)
	if err != nil {
		return nil, err
	}

	return document, nil
}

func ValueFromStringArrayOutput(output pulumi.AnyOutput, index int) pulumi.StringOutput {
	return output.ApplyT(func(arg interface{}) string {
		stringArray := arg.([]interface{})
		return stringArray[index].(string)
	}).(pulumi.StringOutput)
}

func AnyOutputToStringArrayOutput(output pulumi.AnyOutput) pulumi.StringArrayOutput {
	return output.ApplyT(func(arg interface{}) []string {
		stringArray := arg.([]interface{})
		var outputs []string
		for _, s := range stringArray {
			outputs = append(outputs, s.(string))
		}
		return outputs
	}).(pulumi.StringArrayOutput)
}

func StringPtr(input string) *string {
	return &input
}
