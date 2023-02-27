<?php

echo "Today is " . date("Y-m-d") . "\n";

$c = mysqli_connect("127.0.0.1:3306", "user-name", "testtest", "project-name");

if ($c -> connect_errno) {
  echo "Failed to connect to MySQL: " . $c -> connect_error . "\n";
  exit();
} else {
	echo "Connected to MySQL\n";
}

$query = $c->query("SHOW DATABASES");

echo "Databases:\n";

foreach ( $query->fetch_all() as $row) {
  echo $row[0];
  echo "\n";
}
