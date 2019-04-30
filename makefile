init:
	git submodule update --init --recursive
	cd RemoteForensicApplianceFrontend && yarn run build && cp -r dist/* ../web/ && cd ..
	go build -i -o go_build_RemoteForensicAppliance . #gosetup
run:
	go build -i -o go_build_RemoteForensicAppliance . #gosetup
	./go_build_RemoteForensicAppliance #gosetup