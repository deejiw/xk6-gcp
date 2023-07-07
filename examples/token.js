
import { Gcp } from 'k6/x/gcp';

const jsonKey = JSON.parse(open('credentials.json'))

const gcp = new Gcp({
  key: jsonKey,
  scope: ['https://www.googleapis.com/auth/cloud-platform'], // Default value
})
export default function() {
  const accessToken = gcp.getOAuth2AccessToken()
  console.log(accessToken)

  const idToken = gcp.getOAuth2IdToken()
  console.log(idToken)
}
