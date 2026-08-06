package com.personal.repository;

import com.personal.entity.Build;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;

public interface BuildRepository extends JpaRepository<Build, String> {
    List<Build> findByStatus(String status);
}
