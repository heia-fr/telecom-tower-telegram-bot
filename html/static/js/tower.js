var tower = {
  radius: 2,
  margin: 3,
  sep: 1,
  columns: 128,
  rows: 8,
  msgStack: [],
  currentFrame: 0,
  curentMessage: {},

  init: function() {
    this.width = this.columns * (this.radius * 2 + this.sep) - this.sep + 2 * this.margin;
    this.height = this.rows * (this.radius * 2 + this.sep) - this.sep + 2 * this.margin;
    var c = document.getElementById("towerCanvas");
    c.width = this.width;
    c.height = this.height;
    var ctx = c.getContext("2d");

    ctx.beginPath();
    ctx.rect(0,0,this.width,this.height);
    ctx.fillStyle="#333333";
    ctx.fill();

    for (var y=0; y < this.rows; y++) {
      for (var x=0; x < this.columns; x++) {
        ctx.beginPath();
        ctx.arc(
          (this.sep + 2 * this.radius) * x + this.margin + this.radius,
          (this.sep + 2 * this.radius) * y + this.margin + this.radius,
          this.radius, 0, 2*Math.PI);
        ctx.strokeStyle="white";
        ctx.stroke();     
      }
    }
  },

  update_canvas: function(data) {
    var c = document.getElementById("towerCanvas");
    var ctx = c.getContext("2d");

    for (var y=0; y < this.rows; y++) {
      for (var x=0; x < this.columns; x++) {
        ctx.beginPath();
        ctx.arc(
          (this.sep + 2 * this.radius) * x + this.margin + this.radius,
          (this.sep + 2 * this.radius) * y + this.margin + this.radius,
          this.radius, 0, 2*Math.PI);
        var rgb = data[x*this.rows + y]
        ctx.fillStyle = '#' + (0x1000000 + rgb).toString(16).slice(1)
        ctx.fill();     
      }
    }
  },

  roll: function() {
    if (this.currentMessage == null) { // not yet started
      if (this.msgStack.length > 0) {
        this.currentMessage = this.msgStack.shift();
        this.currentFrame = 0;
      } else {
        return;
      }
    }
    var bitmap = this.currentMessage.Matrix.Bitmap;
    var start = this.currentFrame * this.rows;
    var stop = start + this.columns * this.rows;
    if (stop > bitmap.length) {
      stop = bitmap.length;
    }
    var data = bitmap.slice(start, stop);
    this.update_canvas(data);
    this.currentFrame++;
    if ((this.currentFrame + this.columns) * this.rows >= bitmap.length) {
      this.currentFrame = this.currentMessage.Preamble;
    } else if (this.currentFrame == this.currentMessage.Checkpoint) {
      if (this.msgStack.length > 0) {
        this.currentMessage = this.msgStack.shift();
        this.currentFrame = 0;
      }
    }
  }

}

var towerSocket = new ReconnectingWebSocket("wss://telecom-tower.tk/stream");

towerSocket.onmessage = function (event) {
  tower.msgStack.push(JSON.parse(event.data))
}

setInterval(function(){tower.roll();}, 33);

