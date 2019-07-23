init:
	git submodule update --recursive --remote
	cd RemoteForensicApplianceFrontend && yarn run build && cp -r dist/* ../web/ && cd ..
	GOARCH=arm GOOS=linux go build -i -o arm_RemoteForensicAppliance . #gosetup
run:
	./go_build_RemoteForensicAppliance #gosetup