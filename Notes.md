
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




