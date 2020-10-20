/*
    Filter out any strings in the admin-queue
	available variables: queue, body
 */
"";

if (queue !== "admin") {
   body;
   exit;
}

var lines = [];
var input = body.split("\n");
for (var i = 0; i < input.length; i++) {
  if (input[i] === "Hello world2!") continue;
  lines.push(input[i]);
}
lines.join("\n");