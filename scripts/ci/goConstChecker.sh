diff -u <(echo -n) <(goconst -ignore "examples|configmap_source.go" ./...)
if [ $? == 0 ]; then
	echo "No goConst problem"
	exit 0
else
	echo "Has goConst Problem"
	exit 1
fi
