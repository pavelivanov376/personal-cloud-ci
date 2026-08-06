package com.personal.repositories;

import com.personal.entities.BuildEntity;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;

public interface BuildRepository extends JpaRepository<BuildEntity, String> {
    List<BuildEntity> findByJobId(String jobId);
    int countByJobId(String jobId);
}
