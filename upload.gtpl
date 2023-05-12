<html>
<head>
       <title>Upload File</title>
</head>
<body>
<form enctype="multipart/form-data" action="http://192.168.224.132:9090/upload" method="post">
    <input type="file" name="uploadfile" />
    <input type="hidden" name="token" value="{{.}}"/>
    <input type="submit" value="upload" />
</form>
</body>
</html>
