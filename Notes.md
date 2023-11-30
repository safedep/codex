
(import_statement
		name: (dotted_name 
				(identifier) @import_module
		)
	)
	
	(import_from_statement
		 module_name: (dotted_name 
				(identifier) @from_module
		 )
		 name: (dotted_name 
				(identifier) @import_submodule
		)
	)
	
	(import_from_statement
		 module_name: (relative_import
						 (import_prefix) @import_prefix
						(dotted_name (identifier) @from_module)
					   )
		 name: (dotted_name 
				(identifier) @import_submodule
		)
	)



(import_statement
	name: (dotted_name 
    		(identifier) @import_module
    )
)

(import_from_statement
	 module_name: (dotted_name) @from_module
     name: (dotted_name 
    		(identifier) @import_submodule
    )
)

(import_from_statement
	 module_name: (relative_import
     				(import_prefix) @import_prefix
                    (dotted_name (identifier) @from_module)
                   )
     name: (dotted_name 
    		(identifier) @import_submodule
    )
)


(import_statement
	name: (dotted_name 
    		(identifier) @module_name
    )
)

(import_from_statement
	 module_name: (dotted_name) @module_name
     name: (dotted_name 
    		(identifier) @submodule_name
    )
)

(import_from_statement
	 module_name: (relative_import) @module_name
     name: (dotted_name 
    		(identifier) @submodule_name
    )
)



