importScripts("https://cdnjs.cloudflare.com/ajax/libs/BrowserFS/2.0.0/browserfs.js");
importScripts("https://cdn.jsdelivr.net/npm/comlinkjs@3.1.1/umd/comlink.js");
importScripts("https://cdnjs.cloudflare.com/ajax/libs/bluebird/3.5.4/bluebird.min.js");

class GoWorker {
    constructor() {
        this.fs = Promise.promisifyAll(BrowserFS.BFSRequire('fs'));
        console.log("GoWorker constructed");
    }

    //async validate(buffer) {
    //    await this.beforeProcess(buffer);
    //    try {
    //        this.go.argv = ['nexrad.wasm', '-h'];
    //        var st = Date.now();
    //        await this.go.run(this.instance);
    //        console.log('Time taken:', Date.now() - st);
    //
    //        return this.go.exitCode === 0;
    //    } catch (e) {
    //        console.error(e);
    //
    //        return false;
    //    }
    //}

    async extractPage(buffer, argumentsToPass) {
        await this.beforeProcess(buffer);

        var baseArgs = ['nexrad.wasm', '-f', '/radar_file'];
        var userArgs = argumentsToPass.split(' ');
        for (var i = 0; i < userArgs.length; i++) {
            baseArgs.push(userArgs[i])
        }
        console.log(baseArgs)
        this.go.argv = baseArgs;
        var st = Date.now();
        await this.go.run(this.instance);
        console.log('Time taken:', Date.now() - st);

        var filenameToDownload;
        if (baseArgs.includes("png")) {
            filenameToDownload = "radar.png"
        } else if (baseArgs.includes("svg")) {
            filenameToDownload = "radar.svg"
        } else if (baseArgs.includes("svgtest")) {
            filenameToDownload = "TESTradar.svg"
        } else {
            filenameToDownload = "radar.png"
        }

        let contents = await this.fs.readFileAsync(filenameToDownload);
        console.log("after run main:", contents);

        this.fs.unlink("/radar_file", err => {
            console.log("Removed radar_file", err);
            this.fs.unlink(filenameToDownload, err2 => {
                console.log("Removed " + filenameToDownload, err);
            })
        })

        return contents;
    }
    // Write input to /radar_file in browser fs
    async beforeProcess(buffer) {
        // we have to new Go() and create a new instance each time
        // because there are states in the go obj that prevent it from running multiple times
        this.go = new Go();

        if(!this.compiledModule) {
            let result = await WebAssembly.instantiateStreaming(fetch("nexrad.wasm"), this.go.importObject);
            console.log("wasm module compiled!")
            this.compiledModule = result.module; // cache, so that no need to download next time process is called
            this.instance = result.instance;
        } else {
            this.instance = await WebAssembly.instantiate(this.compiledModule, this.go.importObject);
        }

        await this.fs.writeFileAsync('/radar_file', Buffer.from(buffer));
        let contents = await this.fs.readFileAsync('/radar_file');
        console.log(contents);
    }
}

console.log("pre config");
// Configures BrowserFS to use the InMemory file system.
BrowserFS.configure({
    fs: "InMemory"
}, function(e) {
    if (e) {
        // An error happened!
        throw e;
    }
    importScripts('./wasm_exec.js');
    console.log("browserfs initialized!")
    Comlink.expose(GoWorker, self);
});

