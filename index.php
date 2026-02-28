<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <title>PHP基礎：バックエンドへの第一歩</title>
    <style>
        body { font-family: sans-serif; margin: 40px; line-height: 1.6; color: #333; }
        .result { background: #e7f3ff; padding: 20px; border-radius: 10px; border: 1px solid #b3d7ff; margin-top: 20px; }
        input[type="text"] { padding: 10px; width: 250px; border: 1px solid #ccc; border-radius: 4px; }
        button { padding: 10px 20px; background-color: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background-color: #0056b3; }
    </style>
</head>
<body>

    <h1>PHPバックエンド学習：Phase 1</h1>
    <p>フロントエンドから送られたデータを、サーバー側で処理します。</p>
    
    <form action="" method="POST">
        <label for="user_name">あなたのお名前：</label><br>
        <input type="text" id="user_name" name="user_name" placeholder="例：田中太郎" required>
        <button type="submit">サーバーへ送信</button>
    </form>

    <hr>

    <?php
    // --- ここからバックエンド（PHP）の処理 ---
    
    // POSTリクエスト（フォーム送信）があった場合のみ実行
    if ($_SERVER["REQUEST_METHOD"] === "POST") {
        
        // 【重要】フロントから送られてきたデータを受け取る
        // htmlspecialcharsは、不正なスクリプト実行を防ぐ「サニタイズ」という処理です。
        $name = htmlspecialchars($_POST['user_name'], ENT_QUOTES, 'UTF-8');
        
        // サーバー側での加工処理
        $currentTime = date('H時i分s秒');
        $message = "こんにちは、" . $name . "さん！";
        
        // 結果をHTMLとして出力
        echo "<div class='result'>";
        echo "<h2>サーバーからの返信（レスポンス）</h2>";
        echo "<p><strong>{$message}</strong></p>";
        echo "<p>サーバーがこのメッセージを処理した時刻：{$currentTime}</p>";
        echo "<p><small>※これはJavaScriptではなく、PHPがHTMLを生成して返しています。</small></p>";
        echo "</div>";
    }
    ?>

</body>
</html>