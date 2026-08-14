-- char_overlap_ratio is a cheap, multiset-aware character-overlap heuristic
-- used to select a pool of fuzzy-match candidates for the group name search
-- endpoint. For each character in needle, it consumes one matching, not yet
-- consumed occurrence of that character in haystack. The result is
-- (characters matched) / (length of needle), so it is tolerant of
-- misspellings, reordering, and extra characters. Final relevance ranking of
-- the candidate pool is done in Go using Levenshtein distance, since MariaDB
-- has no built-in Levenshtein support on this version.
DELIMITER $$

CREATE FUNCTION char_overlap_ratio(needle VARCHAR(255), haystack VARCHAR(255))
RETURNS FLOAT
DETERMINISTIC
NO SQL
BEGIN
    DECLARE i INT DEFAULT 1;
    DECLARE needleLen INT;
    DECLARE overlap INT DEFAULT 0;
    DECLARE remaining VARCHAR(255);
    DECLARE ch VARCHAR(1);
    DECLARE pos INT;

    SET needle = LOWER(needle);
    SET remaining = LOWER(haystack);
    SET needleLen = CHAR_LENGTH(needle);

    IF needleLen = 0 THEN
        RETURN 0;
    END IF;

    WHILE i <= needleLen DO
        SET ch = SUBSTRING(needle, i, 1);
        SET pos = LOCATE(ch, remaining);
        IF pos > 0 THEN
            SET overlap = overlap + 1;
            SET remaining = CONCAT(SUBSTRING(remaining, 1, pos - 1), SUBSTRING(remaining, pos + 1));
        END IF;
        SET i = i + 1;
    END WHILE;

    RETURN overlap / needleLen;
END$$

DELIMITER ;
