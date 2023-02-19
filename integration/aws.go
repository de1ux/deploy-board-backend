package integration

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/aws/aws-sdk-go/service/s3"
)

func HasS3Buckets(sess *session.Session) (bool, error) {
	svc := s3.New(sess)

	result, err := svc.ListBuckets(nil)
	if err != nil {
		return false, err
	}

	if len(result.Buckets) > 0 {
		return true, nil
	}
	return false, nil
}

func HasCloudfronts(sess *session.Session) (bool, error) {
	svc := cloudfront.New(sess)
	input := &cloudfront.ListDistributionsInput{}

	result, err := svc.ListDistributions(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case cloudfront.ErrCodeInvalidArgument:
				fmt.Println(cloudfront.ErrCodeInvalidArgument, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			return false, err
		}
		return false, err
	}
	if len(result.DistributionList.Items) > 0 {
		return true, nil
	}
	return false, nil
}

func HasRDS(sess *session.Session) (bool, error) {
	svc := elasticbeanstalk.New(sess)
	input := &elasticbeanstalk.DescribeEnvironmentsInput{}

	result, err := svc.DescribeEnvironments(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return false, err
	}
	if len(result.Environments) > 0 {
		svc := elasticbeanstalk.New(sess)
		input := &elasticbeanstalk.DescribeConfigurationSettingsInput{
			ApplicationName: result.Environments[0].ApplicationName,
			EnvironmentName: result.Environments[0].EnvironmentName,
		}

		result, err := svc.DescribeConfigurationSettings(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case elasticbeanstalk.ErrCodeTooManyBucketsException:
					fmt.Println(elasticbeanstalk.ErrCodeTooManyBucketsException, aerr.Error())
				default:
					fmt.Println(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				fmt.Println(err.Error())
			}
			return false, err
		}

		for _, config := range result.ConfigurationSettings {
			for _, option := range config.OptionSettings {
				if *option.OptionName == "HasCoupledDatabase" {
					if *option.Value != "false" {
						return true, err
					}
				}
			}
		}
	}
	return false, nil
}

func HasElasticBeanstalks(sess *session.Session) (bool, error) {
	svc := elasticbeanstalk.New(sess)
	input := &elasticbeanstalk.DescribeEnvironmentsInput{}

	result, err := svc.DescribeEnvironments(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return false, err
	}
	if len(result.Environments) > 0 {
		return true, nil
	}
	return false, nil
}
