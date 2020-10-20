/**
 * NGINX filter the accesslog
 */
if (queue !== "admin") {
	body;
} else {
	var ignores = [
		"access forbidden by rule",
		"Uncaught Error: Load timeout for modules:"
	];
	var lines = [];
        body.split("\n").forEach(function(line) {
		var ok = true;
		ignores.forEach(function(ignore) {
			if (line.indexOf(ignore) !== -1) {
				ok = false;
			}
		});
		if (ok) {
			lines.push(line);
		}
	});
	lines.join("\n");
}
