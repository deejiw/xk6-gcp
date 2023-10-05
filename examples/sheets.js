
import { Gcp } from 'k6/x/gcp';

const jsonKey = JSON.parse(open('credentials.json'))

const gcp = new Gcp({
  key: jsonKey,
  scope: ['https://www.googleapis.com/auth/spreadsheets'], // Default value
})
export default function() {
  const spreadsheetId = "xxx"
  const id = gcp.spreadsheetAppendWithUniqueId(spreadsheetId, 'sheetName', {
    id: 1,
    name: "foo",
  })

  console.log(id)
}
