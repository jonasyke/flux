export namespace main {
	
	export class ModFileResponse {
	    id: number;
	    mod_id: number;
	    filename: string;
	    file_path: string;
	    current_version: string;
	
	    static createFrom(source: any = {}) {
	        return new ModFileResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.mod_id = source["mod_id"];
	        this.filename = source["filename"];
	        this.file_path = source["file_path"];
	        this.current_version = source["current_version"];
	    }
	}

}

