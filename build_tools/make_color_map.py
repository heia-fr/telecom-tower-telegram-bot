import re

print ("colorNames := map[string]string{")
f = open("colors.txt")
for line in f:
	line = line.strip()
	match = re.match(r'(\w*)\s+(#\w+)', line)
	if match:
		key = (match.group(1).lower())
		value = match.group(2)
		print('    "{0}": "{1}",'.format(key,value))
print ("}")
