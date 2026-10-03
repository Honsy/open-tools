/**
 * 多边形包含判断
 * 采用 non-zero winding rule
 */
function isInsidePolygon(points, x, y) {
  var N = points.length;
  var w = 0;

  for (var i = 0, j = N - 1; i < N; i++) {
    var x0 = points[j][0];
    var y0 = points[j][1];
    var x1 = points[i][0];
    var y1 = points[i][1];
    w += windingLine(x0, y0, x1, y1, x, y);
    j = i;
  }
  return w !== 0;
}

function windingLine(x0, y0, x1, y1, x, y) {
  if ((y > y0 && y > y1) || (y < y0 && y < y1)) {
    return 0;
  }
  if (y1 == y0) {
    return 0;
  }
  var dir = y1 < y0 ? 1 : -1;
  var t = (y - y0) / (y1 - y0);
  var x_ = t * (x1 - x0) + x0;

  return x_ > x ? dir : 0;
}

export default {
    isInsidePolygon
}