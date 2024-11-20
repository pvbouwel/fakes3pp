package cmd

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)


const testRegion1 = "tst-1"
const testRegion2 = "eu-test-2"
var backendTestRegions = []string{testRegion1, testRegion2}

var testingBackendsConfig = []byte(fmt.Sprintf(`
# This is a test file check backend-config.yaml if you want to create a configuration
s3backends:
  - region: %s
    credentials:
      file: ../etc/creds/cfc_creds.yaml
    endpoint: http://localhost:5000
  - region: %s
    credentials:
      file: ../etc/creds/otc_creds.yaml
    endpoint: http://localhost:5001
default:  %s
`, testRegion1, testRegion2, testRegion2))


//Set the configurations as expected for the testingbackends
//See testing/README.md for details on testing setup
func setTestingBackendsConfig(t *testing.T) {
	cfg, err :=  getBackendsConfigFromBytes(testingBackendsConfig)
  if err != nil {
    t.Error(err)
    t.FailNow()
  }
  globalBackendsConfig = cfg
}

//This is the testing fixture. It starts an sts and s3 proxy which
//are configured with the S3 backends detailed in testing/README.md.
func testingFixture(t *testing.T) (tearDown func ()(), getToken func(subject string, d time.Duration, tags AWSSessionTags) string){
  //Configure backends to be the testing S3 backends
  setTestingBackendsConfig(t)
	//Given valid server config
  teardownSuiteSTS := setupSuiteProxySTS(t)
  teardownSuiteS3 := setupSuiteProxyS3(t, justProxied)

  //function to stop the setup of the fixture
  tearDownProxies := func () {
    teardownSuiteSTS(t)
    teardownSuiteS3(t)
  }

  _, err := loadOidcConfig([]byte(testConfigFakeTesting))
	if err != nil {
		t.Error(err)
	}
	
	signingKey, err := getTestSigningKey()
	if err != nil {
		t.Error("Could not get test signing key")
		t.FailNow()
	}

  //function to get a valid token that can be exchanged for credentials
  getSignedToken := func(subject string, d time.Duration, tags AWSSessionTags) string {
    token, err := CreateSignedToken(createRS256PolicyTokenWithSessionTags(testFakeIssuer, subject, d, tags), signingKey)
    if err != nil {
      t.Errorf("Could create signed token with subject %s and tags %v: %s", subject, tags, err)
      t.FailNow()
    }
    return token
  }
	

  return tearDownProxies, getSignedToken
}

func getCredentialsFromTestStsProxy(t *testing.T, token, sessionName, roleArn string) aws.Credentials {
	result, err := assumeRoleWithWebIdentityAgainstTestStsProxy(t, token, sessionName, roleArn)
	if err != nil {
		t.Errorf("encountered error when assuming role: %s", err)
	}
  creds := result.Credentials
  awsCreds := aws.Credentials{
    AccessKeyID: *creds.AccessKeyId,
    SecretAccessKey: *creds.SecretAccessKey,
    SessionToken: *creds.SessionToken,
    Expires: *creds.Expiration,
    CanExpire: true,
  }
  return awsCreds
}

//region object is setup in the backends and matches the region name of the backend
func getRegionObjectContent(t *testing.T, region string, creds aws.Credentials) string{
  client := getS3ClientAgainstS3Proxy(t, region, creds)
	
	max1Sec, cancel := context.WithTimeout(context.Background(), 1000 * time.Second)
  var bucketName = "backenddetails"
  var objectKey = "region.txt"
	input := s3.GetObjectInput{
		Bucket: &bucketName,
    Key: &objectKey,
	}
	defer cancel()
	s3ObjectOutput, err := client.GetObject(max1Sec, &input)
	if err != nil {
		t.Errorf("encountered error getting region file for %s: %s", region, err)
	}
  bytes, err := io.ReadAll(s3ObjectOutput.Body)
  if err != nil {
		t.Errorf("encountered error reading region file content for %s: %s", region, err)
	}
  return string(bytes)
}


//Backend selection is done by chosing a region. The enpdoint we use is fixed
//to our testing S3Proxy and therefore the hostname is the same. In each backend
//we have a bucket with the same name and region.txt which holds the actual region
//name which we can use to validate that our request went to the right backend.
func TestMakeSureCorrectBackendIsSelected(t *testing.T) {
  tearDown, getSignedToken := testingFixture(t)
  defer tearDown()
  token := getSignedToken("mySubject", time.Minute * 20, AWSSessionTags{PrincipalTags: map[string][]string{"org": {"a"}}})
  print(token)
  //Given the policy Manager that has roleArn for the testARN
	pm = *NewTestPolicyManagerAllowAll()
  //Given credentials for that role
  creds := getCredentialsFromTestStsProxy(t, token, "my-session", testPolicyAllowAllARN)


  for _, backendRegion := range backendTestRegions {
    regionContent := getRegionObjectContent(t, backendRegion, creds)
    if regionContent != backendRegion {
      t.Errorf("when retrieving region file for %s we got %s", backendRegion, regionContent)
    }
  }
}