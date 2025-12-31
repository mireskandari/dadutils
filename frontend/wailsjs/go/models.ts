export namespace pdf {
	
	export class CombineResult {
	    success: boolean;
	    fileCount: number;
	    pageCount: number;
	    outputSize: number;
	    outputPath: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new CombineResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.fileCount = source["fileCount"];
	        this.pageCount = source["pageCount"];
	        this.outputSize = source["outputSize"];
	        this.outputPath = source["outputPath"];
	        this.error = source["error"];
	    }
	}
	export class CompressionResult {
	    success: boolean;
	    originalSize: number;
	    compressedSize: number;
	    savingsPercent: number;
	    outputPath: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new CompressionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.originalSize = source["originalSize"];
	        this.compressedSize = source["compressedSize"];
	        this.savingsPercent = source["savingsPercent"];
	        this.outputPath = source["outputPath"];
	        this.error = source["error"];
	    }
	}
	export class FileInfo {
	    path: string;
	    name: string;
	    size: number;
	    sizeText: string;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.size = source["size"];
	        this.sizeText = source["sizeText"];
	    }
	}
	export class PDFDocument {
	    id: string;
	    path: string;
	    name: string;
	    pageCount: number;
	    size: number;
	    sizeText: string;
	    pageOrder?: number[];
	
	    static createFrom(source: any = {}) {
	        return new PDFDocument(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.path = source["path"];
	        this.name = source["name"];
	        this.pageCount = source["pageCount"];
	        this.size = source["size"];
	        this.sizeText = source["sizeText"];
	        this.pageOrder = source["pageOrder"];
	    }
	}

}

