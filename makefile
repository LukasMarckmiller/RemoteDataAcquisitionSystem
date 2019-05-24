init:
	git submodule update --recursive --remote
	cd RemoteForensicApplianceFrontend && yarn run build && cp -r dist/* ../web/ && cd ..
	GOARCH=arm go build -i -o go_build_RemoteForensicAppliance . #gosetup
run:
	./go_build_RemoteForensicAppliance #gosetup