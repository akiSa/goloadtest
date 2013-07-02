Config file:
Format: Json

The command list format is as follows:
commandlist: [ { command }, { command } ]
command : { probability: int, iterations: int, sequence: [ { request }, { request } ]
request: { url: string, content-type: string, method: string, post-data: string*, resp: string**, func: []string*** }
USING GROK, %{var} in post-data to go from mapped value (grabbed with grok) to string.

* post-data: the data that will be included in the body of the post request (will be ignored if method = GET)
Format: "{ example: 1, example2: "blah" }" (Assuming it's json)
* resp: How the response will look (If the response does not look like this, it will be ignored. So basically if this is an empty string or if it's left out of the sequence, it will ignore all responses.)
GROK.
%{DAY:var} for day values etc


