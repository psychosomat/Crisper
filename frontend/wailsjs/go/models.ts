export namespace config {
	
	export class Settings {
	    model_name: string;
	    language: string;
	    show_timestamps: boolean;
	    output_dir: string;
	    threads: number;
	    window_frame: string;
	    whisper_path: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model_name = source["model_name"];
	        this.language = source["language"];
	        this.show_timestamps = source["show_timestamps"];
	        this.output_dir = source["output_dir"];
	        this.threads = source["threads"];
	        this.window_frame = source["window_frame"];
	        this.whisper_path = source["whisper_path"];
	    }
	}

}

export namespace hardware {
	
	export class Specs {
	    cpu_threads: number;
	    total_ram_gb: number;
	
	    static createFrom(source: any = {}) {
	        return new Specs(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cpu_threads = source["cpu_threads"];
	        this.total_ram_gb = source["total_ram_gb"];
	    }
	}

}

export namespace models {
	
	export class ModelInfo {
	    name: string;
	    display_name: string;
	    size_gb: number;
	    min_ram_gb: number;
	    speed_factor: number;
	    description: string;
	    filename: string;
	    url: string;
	    sha256: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.display_name = source["display_name"];
	        this.size_gb = source["size_gb"];
	        this.min_ram_gb = source["min_ram_gb"];
	        this.speed_factor = source["speed_factor"];
	        this.description = source["description"];
	        this.filename = source["filename"];
	        this.url = source["url"];
	        this.sha256 = source["sha256"];
	    }
	}

}

export namespace queue {
	
	export class TaskInfo {
	    id: string;
	    file_path: string;
	    file_name: string;
	    status: number;
	    progress: number;
	    error_msg?: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.file_path = source["file_path"];
	        this.file_name = source["file_name"];
	        this.status = source["status"];
	        this.progress = source["progress"];
	        this.error_msg = source["error_msg"];
	    }
	}

}

