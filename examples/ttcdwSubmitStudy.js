define(["Crypto","HepConfig"], function(Crypto,HepConfig) {

	// 视频打点
	function tickerVideo(config) {
		return TickerVideo(config);
	}

	// 文档打点
	function tickerDoc(config) {
		var _this = this;
		var _config = {
			url : "",// 打点记录保存后台地址
			companyCode:"",//三方公司id
			userId : "",// 用户id
			courseId : "",// 课程id
			courseType:"",//课程类型
			resId : "",// 资源id
			resType : "",// 资源类型，文档类的有 ppt pdf
			serverDataName : "tickerData"// 后台接收数据key默认为tickerData
			// onTicker : function(played) {//
			// 打点时回调（played为当前视频已看的对象，请自行console查看格式）
			//
			// }
		}

		if (!config) {
			config = {}
		}

		_this.config = $.extend(_config, config);

		var param = {};
		// console.log(_this.config.companyCode)
		param["companyCode"] = _this.config.companyCode;
		param["userId"] = _this.config.userId;
		param["resId"] = _this.config.resId;
		param["courseId"] = _this.config.courseId;
		param["courseType"] = _this.config.courseType;
		param["resType"] = _this.config.resType;
		param["tickerTime"] = new Date().getTime();

		var serverData = {};
		serverData[_this.config.serverDataName] = JSON.stringify(param);
		// console.log(serverData)
		$.ajax({
			url : _this.config.url,
			type : "post",
			dataType : "json",
			data : serverData,
			success : function(data) {

			}
		})
	}

	// 视频打点抽象类
	function TickerVideo(config) {
		var _this = this;
		var _config = {
			url : "",// 打点记录保存后台地址
			player : "",// 播放器
			companyCode:"",//三方公司id
			userId : "",// 用户id
			courseId : "",// 课程id
			courseType:"",//课程类型
			resId : "",// 资源id
			intervalTime : 30,// 打点间隔时间（秒）
			serverDataName : "tickerData",// 后台接收数据key默认为tickerData
			onTicker : function(played) {// 打点时回调（played为当前视频已看的对象，请自行console查看格式）

			}
		}
		
		clearInterval(_this.countDownId);
		
		if (!config) {
			config = {}
		}

		_this.config = $.extend(_config, config);
		_this.player = _this.config.player;

		// 判断是否是播放状态
		var isPlay = !_this.player.paused();
		if (isPlay) {// 如果是正在播放状态，则直接倒计时开始打点
			TickerVideo.countDown.call(_this);
		}

		// 视频播放回调
		_this.player.on("play", function() {
			TickerVideo.onPlay.call(_this);
		});

		// 视频暂停回调
		_this.player.on("pause", function() {
			TickerVideo.onPause.call(_this);
		});
		
		// 视频结束回调
		_this.player.on("ended", function() {
			TickerVideo.ticker.call(_this);
		});
	}

	// 当用户点击播放，则开始计时打点
	TickerVideo.onPlay = function() {
		var _this = this;
		TickerVideo.countDown.call(_this);
	}

	// 当用户点击暂停，则停止计时
	TickerVideo.onPause = function() {
		var _this = this;
		clearInterval(_this.countDownId);
	}

	// 倒计时
	TickerVideo.countDown = function() {
		var _this = this;
		clearInterval(_this.countDownId);
		_this.countDownId = setInterval(function() {
			TickerVideo.ticker.call(_this);
		}, _this.config.intervalTime * 1000);
	}

	// 打点
	TickerVideo.ticker = function() {
		var _this = this;
		var played = _this.player.played();
		var length = played.length;
		var start = played.start;
		var end = played.end;

		_this.config.onTicker(played);
		// var tickerVideoTimeArray = [];
		// for (var i = 0; i < length; i++) {
		// var tickerVideoTime = {};
		// tickerVideoTime["start"] = played.start(i);
		// tickerVideoTime["end"] = played.end(i);
		// tickerVideoTimeArray.push(tickerVideoTime);
		// }
		var tickerVideoTimeArray = "";
		for (var i = 0; i < length; i++) {
			var timeStr = "";
			if (i == 0) {
				timeStr += played.start(i) + "-" + played.end(i);
			} else {
				timeStr += "," + played.start(i) + "-" + played.end(i);
			}
			tickerVideoTimeArray += timeStr;
		}

		var param = {};
		
		param["companyCode"] = _this.config.companyCode;
		param["userId"] = _this.config.userId;
		param["resId"] = _this.config.resId;
		param["courseId"] = _this.config.courseId;
		param["courseType"] = _this.config.courseType;
		param["tickerTime"] = new Date().getTime();
		param["md5"] = encData(tickerVideoTimeArray);

		var serverData = {};
		serverData[_this.config.serverDataName] = encData(JSON.stringify(param));

		$.ajax({
			url : _this.config.url,
			type : "post",
			dataType : "json",
			data : serverData,
			success : function(data) {

			}
		})
	}
	
	function encrypt (message, key) {
	    var keyHex = Crypto.enc.Utf8.parse(key);
	     var encrypted = Crypto.DES.encrypt(message, keyHex, {
	        mode: Crypto.mode.ECB,
	        padding: Crypto.pad.Pkcs7
	    });
	    return {
	        key: keyHex,
	        value: encrypted.toString()
	    }
	}
	
	function encData(dataStr) {
		var arr = group(dataStr,100);
		var rulArr = [];
		for (var i = 0; i < arr.length; i++) {
			var item = encrypt(arr[i], HepConfig.MD5).value;
			rulArr.push(item);
		}
		return JSON.stringify(rulArr);
	}
	
	function group(string,step) {
        let r = [];
		for(let i = 0, len = string.length; i < len; i+=step) {
			r.push(string.substr(i, step))
		}
		return r;
	}

	return {
		tickerVideo : tickerVideo,
		tickerDoc : tickerDoc
	};
})