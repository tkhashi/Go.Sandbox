export namespace main {
	
	export class ConversionSettings {
	    bitrate: string;
	    sampleRate: string;
	    channels: string;
	
	    static createFrom(source: any = {}) {
	        return new ConversionSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bitrate = source["bitrate"];
	        this.sampleRate = source["sampleRate"];
	        this.channels = source["channels"];
	    }
	}

}

