package com.personal.repository;

import com.personal.entity.BuildEntity;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;

public interface BuildRepository extends JpaRepository<BuildEntity, String> {
    List<BuildEntity> findByStatus(String status);
}
