<?php

namespace "DummyPlugin"

/**
 * This does nothing except create a few version warnings.
 */
trait Logger {
	$log = [];

	public static function Log($message) {
		$log[] = $message;
	}
}