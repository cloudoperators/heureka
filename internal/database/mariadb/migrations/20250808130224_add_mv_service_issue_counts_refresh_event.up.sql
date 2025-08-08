CREATE TABLE mv_service_issue_counts AS
SELECT 
    S.*, 

    COUNT(DISTINCT CASE 
        WHEN IV.issuevariant_rating = 'Critical' 
        THEN CONCAT(CV.componentversion_id, ',', IV.issuevariant_issue_id) 
    END) AS critical_count,

    COUNT(DISTINCT CASE 
        WHEN IV.issuevariant_rating = 'High' 
        THEN CONCAT(CV.componentversion_id, ',', IV.issuevariant_issue_id) 
    END) AS high_count,

    COUNT(DISTINCT CASE 
        WHEN IV.issuevariant_rating = 'Medium' 
        THEN CONCAT(CV.componentversion_id, ',', IV.issuevariant_issue_id) 
    END) AS medium_count,

    COUNT(DISTINCT CASE 
        WHEN IV.issuevariant_rating = 'Low' 
        THEN CONCAT(CV.componentversion_id, ',', IV.issuevariant_issue_id) 
    END) AS low_count,

    COUNT(DISTINCT CASE 
        WHEN IV.issuevariant_rating = 'None' 
        THEN CONCAT(CV.componentversion_id, ',', IV.issuevariant_issue_id) 
    END) AS none_count

FROM 
    Service S
    LEFT JOIN ComponentInstance CI 
        ON S.service_id = CI.componentinstance_service_id
    LEFT JOIN ComponentVersion CV 
        ON CV.componentversion_id = CI.componentinstance_component_version_id
    LEFT JOIN ComponentVersionIssue CVI 
        ON CV.componentversion_id = CVI.componentversionissue_component_version_id
    LEFT JOIN IssueVariant IV 
        ON IV.issuevariant_issue_id = CVI.componentversionissue_issue_id

WHERE 
    S.service_deleted_at IS NULL

GROUP BY 
    S.service_id

ORDER BY  
    critical_count DESC, 
    high_count DESC, 
    medium_count DESC, 
    low_count DESC, 
    none_count DESC

--------

SET GLOBAL event_scheduler = ON;

--------

DELIMITER $$

CREATE EVENT refresh_mv_service_issue_counts
ON SCHEDULE EVERY 1 HOUR
DO
BEGIN
    TRUNCATE mv_service_issue_counts;

    INSERT INTO mv_service_issue_counts
    SELECT 
        S.*, 

        COUNT(DISTINCT CASE 
            WHEN IV.issuevariant_rating = 'Critical' 
            THEN CONCAT(CV.componentversion_id, ',', IV.issuevariant_issue_id) 
        END) AS critical_count,

        COUNT(DISTINCT CASE 
            WHEN IV.issuevariant_rating = 'High' 
            THEN CONCAT(CV.componentversion_id, ',', IV.issuevariant_issue_id) 
        END) AS high_count,

        COUNT(DISTINCT CASE 
            WHEN IV.issuevariant_rating = 'Medium' 
            THEN CONCAT(CV.componentversion_id, ',', IV.issuevariant_issue_id) 
        END) AS medium_count,

        COUNT(DISTINCT CASE 
            WHEN IV.issuevariant_rating = 'Low' 
            THEN CONCAT(CV.componentversion_id, ',', IV.issuevariant_issue_id) 
        END) AS low_count,

        COUNT(DISTINCT CASE 
            WHEN IV.issuevariant_rating = 'None' 
            THEN CONCAT(CV.componentversion_id, ',', IV.issuevariant_issue_id) 
        END) AS none_count

    FROM 
        Service S
        LEFT JOIN ComponentInstance CI 
            ON S.service_id = CI.componentinstance_service_id
        LEFT JOIN ComponentVersion CV 
            ON CV.componentversion_id = CI.componentinstance_component_version_id
        LEFT JOIN ComponentVersionIssue CVI 
            ON CV.componentversion_id = CVI.componentversionissue_component_version_id
        LEFT JOIN IssueVariant IV 
            ON IV.issuevariant_issue_id = CVI.componentversionissue_issue_id

    WHERE 
        S.service_deleted_at IS NULL

    GROUP BY 
        S.service_id

    ORDER BY  
        critical_count DESC, 
        high_count DESC, 
        medium_count DESC, 
        low_count DESC, 
        none_count DESC
END$$

DELIMITER ;
